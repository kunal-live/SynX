package app

import (
	"context"
	"embed"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"synx/internal/config"
	"synx/internal/discovery"
	"synx/internal/events"
	"synx/internal/filesystem"
	"synx/internal/history"
	"synx/internal/identity"
	"synx/internal/observability"
	"synx/internal/pairing"
	"synx/internal/peers"
	"synx/internal/security"
	"synx/internal/server"
	"synx/internal/storage"
	"synx/internal/transfer"
	"synx/internal/transport"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx        context.Context
	mu         sync.RWMutex
	Config     *config.Config
	Identity   *identity.Identity
	DB         *storage.DB
	Bus        *events.Bus
	Metrics    *observability.Metrics
	TrustStore *security.TrustStore
	PeerMgr    *peers.Manager
	Discovery  *discovery.Discovery
	Pairing    *pairing.Service
	Transfer   *transfer.Manager
	History    *history.Service
	Server     *server.Server
	Token      string
	WebFS      embed.FS
}

func New(webFS embed.FS) (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		cfg = config.DefaultConfig()
	}

	id, err := identity.LoadOrCreate(cfg.Storage.DataDirectory, cfg.Device.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize identity: %w", err)
	}

	db, err := storage.Open(cfg.Storage.DatabasePath)
	if err != nil {
		observability.Warn("Failed to open SQLite database: %v. Running in in-memory mode.", err)
	}

	token := security.GenerateToken(8)
	cfg.Security.LocalToken = token

	bus := events.DefaultBus()
	metrics := observability.DefaultMetrics()
	trustStore := security.NewTrustStore(db)
	peerMgr := peers.NewManager(db, bus)
	pairingSvc := pairing.NewService(trustStore, peerMgr, token)
	httpTransport := transport.NewHTTPTransport()
	transferMgr := transfer.NewManager(db, bus, peerMgr, httpTransport, cfg.Transfer.MaxParallelTransfers)
	historySvc := history.NewService(db)

	disc := discovery.New(cfg.Network.DiscoveryPort, cfg.Network.Port, id, token, peerMgr)

	srv := server.New(server.Config{
		Dir:        cfg.Storage.SharedDirectory,
		Token:      token,
		Port:       cfg.Network.Port,
		Web:        webFS,
		Identity:   id,
		TrustStore: trustStore,
		Pairing:    pairingSvc,
		PeerMgr:    peerMgr,
		History:    historySvc,
		DB:         db,
		Bus:        bus,
		AppConfig:  cfg,
	})

	a := &App{
		Config:     cfg,
		Identity:   id,
		DB:         db,
		Bus:        bus,
		Metrics:    metrics,
		TrustStore: trustStore,
		PeerMgr:    peerMgr,
		Discovery:  disc,
		Pairing:    pairingSvc,
		Transfer:   transferMgr,
		History:    historySvc,
		Server:     srv,
		Token:      token,
		WebFS:      webFS,
	}

	return a, nil
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	a.Bus.SetWailsContext(ctx)

	// Start HTTP Server
	if err := a.Server.Start(); err != nil {
		observability.Error("HTTP Server startup failed: %v", err)
	}

	// Start Discovery
	if err := a.Discovery.Start(); err != nil {
		observability.Warn("UDP Discovery startup failed: %v", err)
	}

	observability.Info("SynX Core initialized successfully")
}

func (a *App) Shutdown(ctx context.Context) {
	if a.Discovery != nil {
		a.Discovery.Stop()
	}
	if a.Server != nil {
		_ = a.Server.Stop()
	}
	if a.Transfer != nil {
		a.Transfer.Close()
	}
	if a.PeerMgr != nil {
		a.PeerMgr.Close()
	}
	if a.DB != nil {
		_ = a.DB.Close()
	}
}

// GetState returns system status snapshot for UI
func (a *App) GetState() map[string]any {
	a.mu.RLock()
	defer a.mu.RUnlock()

	peersList := a.PeerMgr.List()
	transfersList := a.Transfer.List()

	ip := localIPv4()
	port := a.Config.Network.Port

	return map[string]any{
		"name":      "SynX",
		"version":   "1.0.0",
		"device_id": a.Identity.DeviceID,
		"platform":  runtime.GOOS,
		"sharedDir": a.Config.Storage.SharedDirectory,
		"token":     a.Token,
		"address":   fmt.Sprintf("%s:%d", ip, port),
		"peers":     peersList,
		"transfers": transfersList,
		"metrics":   a.Metrics.Snapshot(),
	}
}

func (a *App) GetDevices() []peers.Peer {
	return a.PeerMgr.List()
}

func (a *App) GetTransfers() []transfer.Transfer {
	return a.Transfer.List()
}

func (a *App) SendFile(peerID, filePath, destRel string) (*transfer.Transfer, error) {
	return a.Transfer.CreateTransfer(peerID, filePath, destRel)
}

func (a *App) CancelTransfer(id string) error {
	return a.Transfer.Cancel(id)
}

func (a *App) Pair(address, token string) error {
	_, err := a.Pairing.PairDirect(address, token)
	return err
}

func (a *App) GeneratePairingPIN() string {
	return a.Pairing.GeneratePIN()
}

func (a *App) RemovePeer(id string) {
	a.PeerMgr.Remove(id)
	_ = a.TrustStore.RevokePeer(id)
}

func (a *App) ChooseFolder() string {
	if a.ctx == nil {
		return ""
	}
	dir, err := wailsruntime.OpenDirectoryDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title: "Choose SynX shared folder",
	})
	if err != nil || dir == "" {
		return ""
	}
	_ = a.SetSharedDir(dir)
	return dir
}

func (a *App) SetSharedDir(dir string) error {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(abs, 0755); err != nil {
		return err
	}
	return a.Config.SetSharedDir(abs)
}

func (a *App) OpenSharedFolder() {
	if a.ctx != nil {
		wailsruntime.BrowserOpenURL(a.ctx, "file://"+a.Config.Storage.SharedDirectory)
	}
}

func (a *App) GetHistory(limit int) []history.Entry {
	if a.History == nil {
		return []history.Entry{}
	}
	list, _ := a.History.List(limit)
	return list
}

func (a *App) GetSharedFiles(rel string) ([]filesystem.FileItem, error) {
	return filesystem.ScanDirectory(a.Config.Storage.SharedDirectory, rel, filesystem.ScanOptions{
		ExcludeParts: true,
	})
}

func (a *App) DeleteSharedFile(rel string) error {
	target, err := filesystem.ResolveSafePath(a.Config.Storage.SharedDirectory, rel)
	if err != nil {
		return err
	}
	return os.RemoveAll(target)
}

func localIPv4() string {
	ifs, _ := net.Interfaces()
	for _, i := range ifs {
		if i.Flags&net.FlagUp == 0 || i.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, _ := i.Addrs()
		for _, x := range addrs {
			if n, ok := x.(*net.IPNet); ok && n.IP.To4() != nil {
				return n.IP.String()
			}
		}
	}
	return "127.0.0.1"
}
