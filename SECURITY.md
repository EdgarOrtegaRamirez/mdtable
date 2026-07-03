# Security Policy

## Overview

mdtable is a local CLI tool that processes text data. It does not:
- Make network requests
- Store credentials
- Execute arbitrary code
- Process untrusted data in a privileged context

## Input Validation

- All table parsing uses structured state machines, not regex
- Column indices are bounds-checked before access
- File paths are validated before opening
- No shell injection vectors (no `os/exec` with user input)

## Reporting

If you discover a security issue, please open an issue on GitHub.
