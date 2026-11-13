# Roost

**Your Linux dev box, reachable from anywhere, and never lost.**

Roost turns the Linux environments you already have (WSL on a laptop, WSL on a home gaming PC, a VPS, a Raspberry Pi) into **one workspace**. You code in a terminal with Claude Code. You can start at your desk, keep going from your phone through Claude Remote Control while the laptop is closed, carry on from the PC at home, or from a friend's computer if everything else is broken. You never have to remember to sync anything.

> Status: idea and design, iteration 0. Nothing here is built yet. This README describes what we're building so the first person to set it up (the author) can try it and report what broke.

---

## The problem

Claude Code and Remote Control already let you drive a terminal from your phone. But the terminal lives on *one machine*:

- Close the laptop lid → the machine sleeps → the session and its files are gone from your phone.
- Keep a home PC on 24/7 → it works, but someone has to set up WSL, systemd, tmux, SSH, Tailscale, power settings, BIOS, wake-on-LAN, sync and backups by hand. Most people give up halfway.
- The PC loses power, the phone gets stolen and the laptop battery is dead → whatever you did in the last few hours is stuck on a machine you can't reach.

Each piece exists (Tailscale, Syncthing, restic, tmux, wake-on-LAN, `claude remote-control`). Nobody wires them together, least of all on Windows + WSL, where most of the fiddly settings live.

## The idea in one picture

```
                         ┌──────────────── your tailnet (Tailscale) ────────────────┐
                         │                                                           │
   📱 Phone ───── Claude app (Remote Control) ──┐                                    │
   📱 Phone ───── Termius / browser terminal ───┤                                    │
                                                ▼                                    │
   💻 Laptop (WSL) ── VS Code / ssh ──►  🖥️  PRIMARY box  ◄── sync ──►  💻 REPLICA box │
                                        (home PC, WSL)                 (laptop's WSL) │
                                             ▲    │                                   │
                          wake-on-LAN ───────┘    └── encrypted snapshots ──► 🔐 VAULT │
                      🍓 Waker (Raspberry Pi)                 (your own S3 / R2 / B2) │
                         └───────────────────────────────────────────────────────────┘
```

There is **one live copy** of your work (on the *primary*). Everything else is a window into it, a warm copy of it, or a backup of it.

## Concepts

| Term | What it is |
|---|---|
| **Box** | Any Linux environment running the Roost agent: a WSL distro, a VPS, a Pi, a bare Linux machine. |
| **Primary** | The single box where work happens *right now*. Claude sessions, Remote Control, tmux and dev servers run here. Only one box is primary at a time. |
| **Replica** | Any other box. It keeps an up-to-date copy of the workspace (seconds behind) and can be **promoted** to primary. |
| **Vault** | Encrypted, versioned snapshots in storage *you* own (Cloudflare R2, Backblaze B2, S3, Google Drive…). It survives every box being offline. Roost never sees the key. |
| **Waker** | A tiny always-on device at home (Pi, old phone, NAS, router) that can wake a sleeping primary. |
| **Workspace** | What follows you around: your project folders, your Claude history and settings, and selected dotfiles. |
| **Control** | The `roost` CLI and a small web dashboard that runs on every box and is only reachable over your tailnet. |

### Why "one primary" instead of "everything syncs everywhere"

Two machines editing the same repo at the same time is how you get sync conflicts and corrupted `.git` folders. Roost uses a **lease**: the primary holds it, and the other boxes are warm copies. Moving the lease is one command, `roost here`. That rule is what makes "it's always up to date" true, instead of "usually up to date, plus conflict files".

## What follows you (the workspace)

| Synced | Not synced (rebuilt per box) |
|---|---|
| Project folders you pick (e.g. `~/projects`) including `.git` | `node_modules`, `.venv`, `target/`, `dist/`, `.next/`, caches |
| `~/.claude`: sessions, history, settings, skills, plugins | `~/.claude/{cache,debug,shell-snapshots,paste-cache}` |
| Dotfiles you pick: `.gitconfig`, `.bashrc`, `~/.ssh/config` | Docker images and volumes (databases need their own dump/restore) |
| Secrets (`.env`, keys, cloud credentials): **opt-in, vault only, encrypted** | Machine identity: Tailscale node key, SSH host keys |

**The path rule:** every box uses the same username and the same home path (`/home/<you>`). Claude Code stores sessions by folder path, so if the paths match, `claude --resume` on the laptop shows the exact conversation you were having from your phone an hour ago.

## Scenarios

### 1. A normal day

1. **Office, 09:00.** You open the laptop and VS Code connects to `devbox` (the home PC). You open three terminals, and each one is a tmux window *on the PC*. Claude is running in two of them.
2. **Meeting, 11:00.** In the Claude app → Code, both sessions are listed. You voice-note "run the tests and fix what fails." It runs on the PC.
3. **Lunch.** You close the laptop. Nothing stops, because nothing was running on the laptop.
4. **Commute.** You keep working from the phone. You need a folder created and a package installed, so you either ask Claude or open the browser terminal in the Roost dashboard.
5. **Home.** You sit at the PC, open Ubuntu, and run `roost attach`. The same tmux windows show where Claude left off.
6. **Night.** You tap **Sleep** in the dashboard. The PC drops to about 3 W.
7. **Next morning.** You tap **Wake** on the phone (via the Pi). VS Code reconnects. Meanwhile the laptop's replica has been receiving every change, so even its local copy is current.

### 2. Home internet goes down

The dashboard shows `devbox: unreachable`. On the laptop you run:

```
$ roost here
devbox (primary) is unreachable. Last synced 4 min ago.
Promote this laptop to primary anyway? Changes made on devbox after that point
will show up as conflict copies when it comes back. [y/N] y
✓ laptop is primary. Sessions available: claude --resume
```

You work locally all day. When the PC comes back online, the laptop's changes flow back to it. When you get home, run `roost here` on the PC to take the lease back.

### 3. Everything is down: phone stolen, laptop dead, power cut at home

You're at a friend's house.

1. They install Roost (Windows: one installer that creates an isolated WSL distro called `roost`; Linux/macOS: a container).
2. You sign in to your tailnet, then choose **Restore from vault** and type your vault passphrase.
3. Roost pulls the latest snapshot (it was taken a few minutes before the power cut) and installs Claude Code. You log in to Claude and run `roost here`.
4. You work. When you're done: `roost leave --wipe`. This hands the lease to whichever box next comes online, uploads a final snapshot, deletes the distro and logs the friend's machine out of your tailnet.
5. Later, the home PC comes back up after the power cut (the BIOS is set to power on when power returns). It sees it's no longer primary, pulls the newer snapshot and catches up. Same for the laptop when you charge it.
6. For the stolen phone: revoke it in the Tailscale admin and in claude.ai. The phone held no code, so there's nothing else to clean up.

### 4. VPS instead of a home PC

`curl -fsSL roost.sh/install | bash` on a fresh Ubuntu VPS sets it up as a box. Choose "always on" and wake/sleep are skipped. Everything else is identical, and you can move from home PC to VPS, or back, with `roost here`.

## Setup: what the user does vs. what Roost does

```
$ roost init
Detected: Windows 11 + WSL2 (Ubuntu 24.04), Ethernet, 32 GB RAM

  ✓ systemd enabled in WSL
  ✓ tmux, ssh, Tailscale installed       → this box is "devbox"
  ✓ Claude Code installed                → run `claude` once to log in   [ACTION]
  ✓ Remote Control service enabled       → starts on boot
  ✓ Sync: pairing code ABCD-1234         → enter it on your other box
  ✓ Windows: sleep never, hibernate off, Fast Startup off
  ✓ Windows: WSL starts at logon, auto-login configured
  ✓ Network card: wake on magic packet

  ! BIOS — Roost can't change this. On your board (ASUS PRIME B650):
      Advanced → APM → "Restore AC Power Loss" → Power On
      Advanced → APM → "Power On by PCI-E"     → Enabled
      ErP Ready                                → Disabled
    Then run `roost doctor` to verify.                                   [ACTION]
```

| Roost automates | The user has to do |
|---|---|
| WSL config (systemd, memory, idle timeout), packages, tmux, sshd bound to Tailscale only | Install WSL once (or use the installer) |
| Windows power plan, hibernation, Fast Startup, NIC wake settings, autostart task, auto-login | Approve the admin (UAC) prompt |
| Remote Control as a systemd service, restarted if it dies | Log in to Claude once per box (subscription, not API key) |
| Syncthing pairing with the exclude list | Log in to Tailscale once per device |
| Vault schedule (restic), retention, test-restore | Create a storage bucket and choose a passphrase |
| Waker setup on a Pi (one script) | Flash the Pi's SD card |
| `roost doctor`: checks every item above and tells you what's missing | Change BIOS settings (with exact guidance) |

## CLI sketch

```
roost init                 set up this machine as a box
roost doctor               check everything, explain what's broken and how to fix it
roost status               all boxes: primary/replica, online/asleep, last sync, last snapshot
roost here                 make this box primary (graceful handoff if the old one is reachable)
roost attach               attach to the primary's tmux (locally or over ssh)
roost wake <box>           wake a sleeping box via the waker
roost sleep <box>          put a box to sleep (refuses if Claude is mid-task, unless --force)
roost snapshot             push a vault snapshot now
roost restore              rebuild this box from the vault
roost leave [--wipe]       hand off the lease and optionally erase this box
roost dash                 open the web dashboard (tailnet only)
```

## Where data lives

| Data | Lives on | Leaves your control? |
|---|---|---|
| Code, files, git history | Primary, plus replicas | No: box-to-box over your tailnet |
| Encrypted snapshots | Your bucket | Encrypted before upload; the key is only your passphrase |
| Claude conversation (while Remote Control is connected) | Anthropic | Yes, the transcript, as with any Remote Control use |
| Box list, lease, health | A small state file synced between boxes | No |
| Roost servers | **None.** There is no Roost account or cloud | — |

## Security model

- Nothing listens on the public internet. SSH and the dashboard bind to the Tailscale interface only. Remote Control connects outbound to Anthropic.
- Authentication is your tailnet identity, plus SSH keys or passkeys for the dashboard terminal.
- Lost device: remove it from the tailnet. Roost shows each box's last check-in so you can spot devices you've forgotten about.
- Secrets never travel through Syncthing unless you opt in. By default they live only in the encrypted vault and on the primary.
- `roost leave --wipe` is what you use on a machine you don't own.

## Hard problems (honest list)

1. **BIOS settings cannot be automated.** The best we can do is detect the motherboard and show the exact menus.
2. **Wake-on-LAN needs Ethernet and a device at home.** Over Wi-Fi it rarely works. A smart plug plus "power on after AC loss" is the fallback.
3. **Promoting a box while the primary is unreachable can produce conflicts.** Syncthing keeps both versions, and Roost has to show them clearly.
4. **A live Remote Control session can't move between boxes.** The conversation can, via `claude --resume`; the running process can't. Handoff means: finish or pause, sync, resume on the new box.
5. **Databases and Docker volumes** don't sync safely as files. They need per-project dump/restore hooks.
6. **Big workspaces:** dozens of GB of worktrees and archives. A first sync over home upload speed can take hours; the vault should exclude archives by default.
7. **Windows quirks:** WSL shuts down idle distros, mirrored networking, Windows Update restarts. `roost doctor` has to catch these.

## What already exists

- **Tailscale**: the network. Roost builds on it rather than replacing it.
- **Syncthing**: box-to-box sync. **restic**: the vault.
- **UpSnap**: a wake-on-LAN web UI. **Pocketdev**: a single Linux devbox for phone SSH. **Wolfpack / Kunai**: driving coding agents from a phone.
- None of these combine Windows/WSL setup, power recovery, wake/sleep, single-primary sync, an off-site vault and Claude Remote Control into one install. That combination is Roost.

## Roadmap

| Iteration | Goal | Shape |
|---|---|---|
| **0 (now)** | Do it by hand on the author's own laptop + home PC + Pi. Write down every step and every failure. | This README + `NOTES.md` |
| **1** | Scripts: `roost-init.sh` (WSL side), `roost-windows.ps1` (Windows side), `roost-waker.sh` (Pi), `roost doctor` | Bash + PowerShell |
| **2** | Lease + `roost here` / `status` / `wake` / `sleep`, Syncthing pairing | Bash, still small |
| **3** | Vault (restic) + `restore` + `leave --wipe`: the friend's-laptop scenario works | — |
| **4** | Single binary (Go: one file for Windows, Linux x64 and ARM) + web dashboard with browser terminal | Go + small web UI |
| **5** | Windows installer (.exe) that creates the WSL distro for people who have never used WSL | — |

## Open questions

- Is Tailscale a hard requirement, or do we support plain WireGuard / Headscale from the start?
- Should the vault also auto-commit WIP to a `roost/wip` git branch as a second safety net?
- Do we support macOS as a *box*, or only as a client?
- Name: **Roost** is a working title ("your dev box comes home to roost"). Alternatives: *Hearth*, *Homeport*, *Tether*, *Wayhome*. Check for conflicts on GitHub, npm and domains before the first release.

## License

TBD (MIT or Apache-2.0 likely).
