# Roost

**Close your laptop and keep coding from your phone.** Roost turns a machine you already have (a Windows gaming PC with WSL, a VPS, a Raspberry Pi) into an always-on Linux dev box. You reach it from your phone through Claude Code Remote Control, and from your laptop through SSH or VS Code.

One command sets it up. A second command checks it.

```bash
curl -fsSL https://raw.githubusercontent.com/CemOlmez/roost/main/install.sh | bash
```

> **Status: 0.1, early.** Box setup and checks work. Sync between machines, an encrypted off-site backup and wake/sleep from your phone are planned (see [Roadmap](#roadmap)). The full design is in [docs/DESIGN.md](docs/DESIGN.md).

---

## How it works

```
 📱 Phone ── Claude app (Remote Control) ─┐
 📱 Phone ── Termius (SSH) ───────────────┤      private network (Tailscale)
 💻 Laptop ── VS Code / roost attach ─────┼──────────────►  🖥️ BOX  (home PC / VPS / Pi)
                                          │                   • your code and terminals live here
                                          │                   • Claude Remote Control runs 24/7
                                          │                   • comes back by itself after reboots and power cuts
```

Your work lives on **one machine, the box**. Your laptop and phone are windows into it, so there's nothing to keep in sync: close the laptop and the box keeps running.

There are two roles:

| Role | What it is | What `roost init` sets up |
|---|---|---|
| **box** | The machine that runs your work 24/7 | Tailscale + SSH, Claude Code, Remote Control as a background service, and on Windows: WSL starts at logon, no sleep, no hibernate, wake-on-LAN |
| **client** | A laptop you connect from | Tailscale + SSH, Claude Code, and `roost attach` to jump into the box's terminal |

---

## Before you start: what only you can do

Roost automates everything it can. These steps need you, because they involve accounts, passwords or hardware:

**Accounts (once)**
- [ ] A **Claude Pro, Max, Team or Enterprise** subscription. Remote Control does not work with an API key.
- [ ] A free **[Tailscale](https://tailscale.com)** account. It's the private network that connects your devices; nothing is exposed to the internet.

**On your phone**
- [ ] **Claude** app: sign in. This is where you'll text or voice-note your box.
- [ ] **Tailscale** app: sign in with the same account.
- [ ] **Termius** (optional): a terminal for running commands yourself.

**On the box machine**
- [ ] Linux, or **Windows 10/11 with WSL2 + Ubuntu**. To get WSL, open PowerShell and run `wsl --install`, then reboot.
- [ ] If it's a home PC: an **Ethernet cable**, which is needed for wake-on-LAN later and is more reliable in general.
- [ ] Your **Linux password** for `sudo`, and **Windows admin** rights (you'll get one UAC prompt).

`roost init` will also ask you to do two things that no software can do for you, and shows exact instructions for both:
- **Windows auto-login** (Sysinternals Autologon), so the PC comes back up after a power cut or an update.
- **BIOS settings**: "Restore on AC power loss → Power On", "Wake on LAN → Enabled" and "ErP → Disabled".

---

## Setup

### 1. On the box (e.g. your home PC, in Ubuntu)

```bash
curl -fsSL https://raw.githubusercontent.com/CemOlmez/roost/main/install.sh | bash
```

It asks what the machine is (choose **box**) and what to call it (default `devbox`). Then it checks each item and offers to fix what's missing:

```
  ✓ WSL runs systemd
  ✓ Base packages: tmux git curl python3
  ✗ Tailscale connected as 'devbox' with Tailscale SSH on
  Fix this now? [Y/n]
  ...
  ! BIOS: power on after power loss + wake on LAN   [your action]
      Motherboard: ASUS PRIME B650. Reboot, press Del/F2 ...
```

Re-run `roost init` as often as you like. It only touches what isn't set up yet.

### 2. On your laptop (in Ubuntu/WSL)

Run the same command and choose **client**. Then:

```bash
roost attach        # opens your persistent terminal on the box
```

With VS Code, install the **Remote - SSH** extension and connect to `devbox`.

### 3. On your phone

Open the Claude app → **Code**. Your box (`devbox`) is listed: open it and type or talk.

### 4. Test it

- [ ] Start something in `roost attach`, close the laptop, and continue from the phone.
- [ ] Restart the box PC: Windows logs in, WSL starts, and the box shows up in the Claude app again by itself.
- [ ] Pull the power plug, then plug it back in: the PC turns on by itself (BIOS setting).

---

## Commands

```
roost init       set up / repair this machine
roost doctor     check everything, change nothing
roost status     this machine, your devices, Remote Control
roost attach     your persistent terminal (tmux); over SSH on a client
roost rc attach  see the Remote Control screen (QR code, sessions)
roost rc logs    Remote Control service logs
```

## What Roost changes on your machine

Nothing is hidden. Every change is one of these:

| Where | Change |
|---|---|
| Ubuntu | `/etc/wsl.conf` → `systemd=true`; installs tmux, git, curl, python3, Tailscale, Claude Code; `loginctl enable-linger`; user service `~/.config/systemd/user/roost-rc.service` |
| Windows (box only) | `%USERPROFILE%\.wslconfig` → `vmIdleTimeout=-1` (backup saved); scheduled task **RoostKeepAlive**; `powercfg`: no sleep on AC, hibernate off; Ethernet adapter: wake on magic packet |
| Config | `~/.config/roost/config` |

To undo: `systemctl --user disable --now roost-rc`, delete the **RoostKeepAlive** task in Task Scheduler, and run `powercfg /hibernate on`.

## Roadmap

| Version | Adds |
|---|---|
| **0.1** (now) | Box + client setup, `doctor`, Remote Control service, Windows always-on |
| 0.2 | `roost wake` / `roost sleep` from your phone through a Raspberry Pi |
| 0.3 | Laptop keeps a live copy (Syncthing); `roost here` makes the laptop the main machine when the box is unreachable |
| 0.4 | Encrypted off-site vault (restic), so you can restore your whole workspace on any computer with `roost restore` |
| 0.5 | Web dashboard + one-click Windows installer |

## License

MIT
