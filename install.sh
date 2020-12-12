#!/bin/bash

set -eu

check_go() {
  if ! command -v go &> /dev/null; then
    printf "\nGo could not be found, please install it first\n\n"
    return 1
  fi

  export GOPATH=$HOME
  export GOBIN=$GOPATH/bin
  mkdir -p "$GOBIN"
}

compile_bin_files() {
  cd "$INSTALL_SCRIPT_DIR/cmd/pocket2rm-setup"
  go get
  go build main.go

  cd "$INSTALL_SCRIPT_DIR/cmd/pocket2rm"
  go get
  GOOS=linux GOARCH=arm GOARM=7 go build -o pocket2rm.arm

  cd "$INSTALL_SCRIPT_DIR/cmd/pocket2rm-reload"
  go get
  GOOS=linux GOARCH=arm GOARM=7 go build -o pocket2rm.arm

  printf "\n\n"
  "$INSTALL_SCRIPT_DIR/cmd/pocket2rm-setup/main"

  printf "\npocket2rm successfully compiled\n\n"
}

copy_bin_files_to_remarkable() {
  cd "$INSTALL_SCRIPT_DIR"
  scp "$HOME/.pocket2rm" root@"$REMARKABLE_IP":/home/root/.
  scp cmd/pocket2rm/pocket2rm.arm root@"$REMARKABLE_IP":/home/root/.
  scp cmd/pocket2rm-reload/pocket2rm-reload.arm root@"$REMARKABLE_IP":/home/root/.
}

copy_service_files_to_remarkable() {
  cd "$INSTALL_SCRIPT_DIR"
  scp cmd/pocket2rm/pocket2rm.service root@"$REMARKABLE_IP":/etc/systemd/system/.
  scp cmd/pocket2rm-reload/pocket2rm-reload.service root@"$REMARKABLE_IP":/etc/systemd/system/.
}

register_and_run_service_on_remarkable() {
  ssh root@"$REMARKABLE_IP" systemctl enable pocket2rm-reload
  ssh root@"$REMARKABLE_IP" systemctl start pocket2rm-reload
}

INSTALL_SCRIPT_DIR=""
REMARKABLE_IP=""

main() {
  INSTALL_SCRIPT_DIR=$(pwd)

  printf "\n\n"
  read  -r -p "Enter your Remarkable IP address [10.11.99.1]: " REMARKABLE_IP
  REMARKABLE_IP=${REMARKABLE_IP:-10.11.99.1}
  
  if [ ! -f "$HOME/.pocket2rm" ]; then
    check_go
    compile_bin_files
    copy_bin_files_to_remarkable
  fi

  copy_service_files_to_remarkable
  register_and_run_service_on_remarkable

  printf "\n\npocket2rm successfully installed on your Remarkable\n\n"
}

main