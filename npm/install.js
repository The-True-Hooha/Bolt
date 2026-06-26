#!/usr/bin/env node
'use strict';

const https = require('https');
const fs = require('fs');
const path = require('path');
const os = require('os');
const { execSync } = require('child_process');

const VERSION = '0.1.5';
const REPO = 'The-True-Hooha/Bolt';
const BIN_DIR = path.join(__dirname, 'bin');

function getPlatformInfo() {
  const platform = process.platform;
  const arch = process.arch;

  let os_name, os_arch, ext, archiveExt;

  if (platform === 'darwin') {
    os_name = 'darwin';
    archiveExt = 'tar.gz';
    ext = '';
  } else if (platform === 'linux') {
    os_name = 'linux';
    archiveExt = 'tar.gz';
    ext = '';
  } else if (platform === 'win32') {
    os_name = 'windows';
    archiveExt = 'zip';
    ext = '.exe';
  } else {
    throw new Error(`Unsupported platform: ${platform}`);
  }

  if (arch === 'x64') {
    os_arch = 'amd64';
  } else if (arch === 'arm64') {
    os_arch = 'arm64';
  } else {
    throw new Error(`Unsupported architecture: ${arch}`);
  }

  const baseName = `bolt-fm_${VERSION}_${os_name}_${os_arch}`;
  const archiveName = `${baseName}.${archiveExt}`;
  const binaryName = `${baseName}${ext}`;
  const url = `https://github.com/${REPO}/releases/download/v${VERSION}/${archiveName}`;

  return { url, archiveName, binaryName, ext };
}

function download(url, dest) {
  return new Promise((resolve, reject) => {
    const follow = (u) => {
      https.get(u, (res) => {
        if (res.statusCode === 301 || res.statusCode === 302) {
          follow(res.headers.location);
          return;
        }
        if (res.statusCode !== 200) {
          reject(new Error(`Download failed with status ${res.statusCode}: ${u}`));
          return;
        }
        const file = fs.createWriteStream(dest);
        res.pipe(file);
        file.on('finish', () => file.close(resolve));
        file.on('error', reject);
      }).on('error', reject);
    };
    follow(url);
  });
}

async function install() {
  const { url, archiveName, binaryName, ext } = getPlatformInfo();

  if (!fs.existsSync(BIN_DIR)) {
    fs.mkdirSync(BIN_DIR, { recursive: true });
  }

  const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), 'bolt-fm-install-'));
  const archivePath = path.join(tmpDir, archiveName);
  const finalBinary = path.join(BIN_DIR, `bolt-fm${ext}`);

  console.log(`Downloading bolt-fm v${VERSION} from ${url} ...`);
  await download(url, archivePath);

  console.log('Extracting...');
  if (archiveName.endsWith('.tar.gz')) {
    execSync(`tar -xzf "${archivePath}" -C "${tmpDir}"`);
  } else if (archiveName.endsWith('.zip')) {
    // Use PowerShell on Windows
    execSync(
      `powershell -NoProfile -Command "Expand-Archive -Path '${archivePath}' -DestinationPath '${tmpDir}' -Force"`
    );
  }

  const extractedBinary = path.join(tmpDir, binaryName);
  if (!fs.existsSync(extractedBinary)) {
    throw new Error(`Expected binary not found after extraction: ${extractedBinary}`);
  }

  fs.copyFileSync(extractedBinary, finalBinary);

  if (ext === '') {
    fs.chmodSync(finalBinary, 0o755);
  }

  // Cleanup
  fs.rmSync(tmpDir, { recursive: true, force: true });

  console.log(`bolt-fm installed to ${finalBinary}`);
}

install().catch((err) => {
  console.error('bolt-fm install failed:', err.message);
  process.exit(1);
});
