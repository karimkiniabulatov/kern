#!/bin/bash
set -e

echo "Installing kern systemd service..."

# Создаем systemd service файл
SERVICE_FILE="/etc/systemd/system/kern.service"

if [ ! -w "/etc/systemd/system" ]; then
    echo "❌ Need root privileges to install systemd service"
    echo "Please run: sudo $0"
    exit 1
fi

cat > $SERVICE_FILE << EOF
[Unit]
Description=kern System Monitoring Daemon
After=network.target
Wants=network.target

[Service]
Type=simple
User=$USER
ExecStart=/usr/local/bin/kern --remote 28126
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
Environment=HOME=/home/$USER

[Install]
WantedBy=multi-user.target
EOF

# Перезагружаем systemd и включаем сервис
systemctl daemon-reload
systemctl enable kern.service
systemctl start kern.service

echo "✅ kern systemd service installed and started"
echo "🔧 Management commands:"
echo "   sudo systemctl status kern"
echo "   sudo systemctl stop kern" 
echo "   sudo systemctl start kern"
echo "   sudo journalctl -u kern -f"