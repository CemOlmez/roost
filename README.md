# Roost

[![CI](https://github.com/CemOlmez/roost/actions/workflows/ci.yml/badge.svg)](https://github.com/CemOlmez/roost/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/CemOlmez/roost)](https://github.com/CemOlmez/roost/releases)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

**Close your laptop and keep coding from your phone.**

Roost turns a machine you already own (a Windows gaming PC with WSL, a VPS, a Raspberry Pi) into an **always-on Linux dev box**. You reach it from your phone with Claude Code Remote Control, and from your laptop with SSH or VS Code. It comes back by itself after reboots and power cuts.

```bash
curl -fsSL https://raw.githubusercontent.com/CemOlmez/roost/main/install.sh | sh
```

One command installs it, `roost init` sets the machine up, and `roost doctor` checks it any time.

> **Status: v0.1, early.** Setting up a box and a client works. Syncing between machines, an encrypted off-site vault and waking the PC from your phone are on the [roadmap](#roadmap). If something breaks on your machine, please [open an issue](https://github.com/CemOlmez/roost/issues).

---

## What it is

Roost is a **small program (one file, written in Go) that you install on your own machines**. It is not a cloud service: there's no Roost account and no Roost server, and your code never passes through anything we run.

It sets up and keeps healthy the tools that make an always-on dev box work: **Tailscale** (private network), **tmux** (terminals that survive disconnects), **Claude Code Remote Control** (your phone drives the terminal), **systemd** (starts everything on boot), and on Windows the power, startup and wake-on-LAN settings.

```
 📱 Phone ── Claude app (Remote Control) ─┐
 📱 Phone ── Termius (SSH) ───────────────┤      private network (Tailscale)
 💻 Laptop ── VS Code / roost attach ─────┴─────────────►  🖥️ BOX  (home PC / VPS / Pi)
                                                            • your code and terminals live here
                                                            • Claude Remote Control runs 24/7
                                                            • back by itself after reboots and power cuts
```

Your work lives on **one machine, the box**. Your laptop and phone are windows into it, so there is nothing to sync: close the laptop and the box keeps going.

| Role | What it is | What `roost init` sets up |
|---|---|---|
| **box** | The machine that runs your work 24/7 | Tailscale + SSH, Claude Code, Remote Control as a background service, and on Windows: WSL starts at logon, no sleep, Fast Startup off, wake-on-LAN |
| **client** | A laptop you connect from | Tailscale + SSH, Claude Code, `roost attach` to jump into the box's terminal |

## Before you start

Roost automates everything it can. These things need you, because they involve accounts, passwords or hardware:

**Accounts**
- [ ] **Claude Pro, Max, Team or Enterprise.** Remote Control doesn't work with an API key.
- [ ] A free **[Tailscale](https://tailscale.com)** account. Nothing gets exposed to the internet.

**Phone**
- [ ] **Claude** app, signed in.
- [ ] **Tailscale** app, signed in with the same account.
- [ ] **Termius** (optional), for running commands yourself.

**The box machine**
- [ ] Linux, or **Windows 10/11 with WSL2 + Ubuntu**. To get WSL, open PowerShell as administrator, run `wsl --install`, then reboot.
- [ ] For a home PC: an **Ethernet cable**. Wake-on-LAN needs it, and it's more reliable anyway.
- [ ] Your Linux password (for `sudo`) and Windows administrator rights (you'll get one UAC prompt).

`roost init` also asks for two things software can't do, with exact instructions for each:
- **Windows auto-login**, so the PC comes back after a power cut or an update.
- **BIOS settings**: power on after a power loss, and wake-on-LAN.

## Setup

**1. On the box** (e.g. your home PC, inside Ubuntu):

```bash
curl -fsSL https://raw.githubusercontent.com/CemOlmez/roost/main/install.sh | sh
```

Choose **box** and accept the name `devbox`. Roost then goes through its checklist:

```
  ✓ WSL runs systemd (needed for background services)
  ✓ Base packages installed: tmux, git, curl
  ✗ Tailscale connected as "devbox" with Tailscale SSH on
      Tailscale SSH is off
  Fix this now? [Y/n]
  ...
  ! BIOS: PC powers on after a power cut, wake-on-LAN enabled  [your action]
      Motherboard: ASUS PRIME B650-PLUS. Reboot and press Del or F2 ...
```

Re-run `roost init` whenever you like. It only touches what isn't set up yet.

**2. On your laptop** (in Ubuntu/WSL): run the same command and choose **client**. Then:

```bash
roost attach      # your persistent terminal, on the box
```

In VS Code, install the **Remote - SSH** extension and connect to `devbox`.

**3. On your phone:** open the Claude app → **Code**. `devbox` is listed: open it, then type or talk.

**4. Prove it works:**
- [ ] Start something in `roost attach`, close the laptop, and carry on from the phone.
- [ ] Restart the box: Windows logs in, WSL starts, and `devbox` reappears in the Claude app on its own.
- [ ] Pull the box's power cable and plug it back in: the PC starts by itself.

## Commands

| Command | Does |
|---|---|
| `roost init` | Set up or repair this machine. Flags: `--role`, `--name`, `--box`, `--rc-dir`, `--yes` |
| `roost doctor` | Check everything, change nothing. Exits 1 if anything needs attention; `--json` for scripts |
| `roost status` | This machine, your devices, Remote Control. `--json` available |
| `roost attach` | Your persistent terminal (tmux), over SSH when run on a client |
| `roost rc attach` | See the Remote Control screen (QR code, sessions). Detach: `Ctrl-b d` |
| `roost rc logs` / `start` / `stop` / `restart` | Manage the Remote Control service |
| `roost confirm bios` | Mark the BIOS step done (Roost can't detect BIOS settings) |

## What Roost changes on your machine

Nothing is hidden, and every change can be undone.

| Where | Change | Undo |
|---|---|---|
| Ubuntu | `/etc/wsl.conf` → `[boot] systemd=true` | remove the line |
| Ubuntu | installs tmux, git, curl, Tailscale, Claude Code | `apt remove …` |
| Ubuntu | Tailscale: `--ssh --hostname=<name> --operator=<you>` | `sudo tailscale set --ssh=false` |
| Ubuntu (box) | `loginctl enable-linger`, user service `~/.config/systemd/user/roost-rc.service` | `systemctl --user disable --now roost-rc` |
| Windows (box) | `%USERPROFILE%\.wslconfig` → `vmIdleTimeout=-1` (backup: `.wslconfig.roost-backup`) | restore the backup |
| Windows (box) | scheduled task **RoostKeepAlive**, no sleep on AC, hibernate off, Ethernet wake on magic packet | delete the task; `powercfg /hibernate on` |
| Roost | `~/.config/roost/` | delete it |

## Roadmap

| Version | Adds |
|---|---|
| **0.1** | Box + client setup, `doctor`, always-on Remote Control, Windows always-on |
| 0.2 | `roost wake` / `roost sleep` from your phone, through a Raspberry Pi at home |
| 0.3 | Your laptop keeps a live copy (Syncthing); `roost here` makes the laptop the main machine when the box is down |
| 0.4 | Encrypted off-site vault (restic): `roost restore` your whole workspace on any computer |
| 0.5 | Web dashboard + one-click Windows installer |

The full vision and scenarios are in [docs/DESIGN.md](docs/DESIGN.md), and how the code works is in [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md).

## Contributing

Bug reports from real setups are the most valuable thing right now. Include `roost doctor --json` output. See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

[MIT](LICENSE)
