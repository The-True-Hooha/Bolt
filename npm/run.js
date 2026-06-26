#!/usr/bin/env node
'use strict';

const { spawnSync } = require('child_process');
const path = require('path');
const fs = require('fs');

const isWindows = process.platform === 'win32';
const binaryName = isWindows ? 'bolt-fm.exe' : 'bolt-fm';
const binaryPath = path.join(__dirname, 'bin', binaryName);

if (!fs.existsSync(binaryPath)) {
  console.error(
    `bolt-fm binary not found at ${binaryPath}.\n` +
    'Try reinstalling: npm install @the-true-hooha/bolt-fm'
  );
  process.exit(1);
}

const result = spawnSync(binaryPath, process.argv.slice(2), {
  stdio: 'inherit',
  windowsHide: false,
});

if (result.error) {
  console.error('Failed to run bolt-fm:', result.error.message);
  process.exit(1);
}

process.exit(result.status ?? 1);
