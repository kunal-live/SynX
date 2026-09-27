const fs = require('fs');
const path = require('path');

const srcDir = path.resolve(__dirname, '../out');
const destDesktop = path.resolve(__dirname, '../../desktop/frontend');
const destWeb = path.resolve(__dirname, '../../web');

function cleanDirExcept(dir, preserveNames = []) {
  if (!fs.existsSync(dir)) return;
  const entries = fs.readdirSync(dir);
  for (const entry of entries) {
    if (preserveNames.includes(entry)) continue;
    fs.rmSync(path.join(dir, entry), { recursive: true, force: true });
  }
}

function copyDirRecursive(src, dest) {
  if (!fs.existsSync(src)) return;
  if (!fs.existsSync(dest)) fs.mkdirSync(dest, { recursive: true });

  const entries = fs.readdirSync(src, { withFileTypes: true });
  for (const entry of entries) {
    const srcPath = path.join(src, entry.name);
    const destPath = path.join(dest, entry.name);

    if (entry.isDirectory()) {
      copyDirRecursive(srcPath, destPath);
    } else {
      fs.copyFileSync(srcPath, destPath);
    }
  }
}

if (fs.existsSync(srcDir)) {
  console.log(`[SynX Sync] Copying Next.js build from ${srcDir} to ${destDesktop} and ${destWeb}...`);
  cleanDirExcept(destDesktop, ['wailsjs']);
  cleanDirExcept(destWeb, []);
  copyDirRecursive(srcDir, destDesktop);
  copyDirRecursive(srcDir, destWeb);
  console.log(`[SynX Sync] Next.js build successfully mirrored to desktop and web distributions.`);
} else {
  console.error(`[SynX Sync] Source directory ${srcDir} does not exist. Did 'next build' run?`);
}
