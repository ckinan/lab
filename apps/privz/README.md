# privz

A CLI tool to inspect connected hardware devices and check camera and microphone privacy states on Linux. Instead of checking desktop settings or htop, `privz` scans system buses and inspects `/proc/*/fd` to tell you exactly what peripherals are attached and which processes are actively streaming audio or video.

## Installation

```bash
go install github.com/ckinan/lab/apps/privz@latest
```

## Usage

Run `privz` without arguments to print a snapshot of all connected devices and media states:

```bash
privz
```

### Configuration (Optional)

If you want to filter out noisy motherboard virtual devices or customize sorting, you can create a configuration file. `privz` looks for configuration in two places (in this order):

1. `~/.privz.yaml`
2. `~/.config/privz/config.yaml`

Example structure:

```yaml
# Categories to exclude entirely from output
ignore_categories:
  - "INPUT/OTHER"

# Substrings to skip matching specific device names
ignore_keywords:
  - "Power Button"
  - "Sleep Button"
  - "PC Speaker"

# Default sort order ('category' or 'name')
sort_by: category
```

### Output Example

```text
=== PRIVZ HARDWARE & PRIVACY RADAR ===

--- SPECIAL PRIVACY SECTION (CAMERA & MICROPHONE) ---
DEVICE        STATUS                     CONNECTION      ACTIVE PROCESS
Camera        [ON] (STREAMING ACTIVE)    Connected       chrome (PIDs: [8192])
Microphone    [OFF]                      Connected       -

--- ALL CONNECTED HARDWARE DEVICES ---
CATEGORY          DEVICE NAME                      SYSTEM PATH                     DETAILS
[AUDIO/MIC]       HDA-Intel - HDA Intel PCH        /proc/asound/cards              ALSA Sound Card
[DISPLAY]         card0-DP-1                       /sys/class/drm/card0-DP-1       DRM Display Connector (Connected)
[KEYBOARD]        USB usb keyboard                 sysrq kbd leds event4           evdev input device
[MOUSE/TOUCH]     Logitech MX Master 3             sysrq kbd leds mouse0 event8    evdev input device
```
