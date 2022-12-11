build() {
  GOOS=linux go build || exit 1
}

build_turris() {
  GOARCH=arm GOARM=7 build
}

redeploy() {
  build
  ssh $DDNSSERV '/etc/init.d/ddnsd stop'
  scp go-ddnsd $DDNSSERV:/srv/ddnsd
  ssh $DDNSSERV '/etc/init.d/ddnsd start'
  ssh_logs
}

redeploy_turris() {
  build_turris
  DDNSSERV=turris
  ssh $DDNSSERV '/etc/init.d/ddnsd stop'
  scp go-ddnsd $DDNSSERV:/srv/ddnsd
  ssh $DDNSSERV '/etc/init.d/ddnsd start'
  ssh_logs
}

ssh_logs() {
    ssh $DDNSSERV 'logread -f'
}

redeploy_service() {
  scp ./etc/init.d/ddnsd $DDNSSERV:/etc/init.d/
}

test_dns() {
  curl -X GET --location "http://dyn.jkl.mn:8080/nic/update?hostname=d1.dyn.jkl.mn&token=PASSWORD"
  dig d1.dyn.jkl.mn
}

test_dns_loc() {
  curl -X GET --location "http://localhost:8080/nic/update?hostname=d1.dyn.jkl.mn&token=PASSWORD"
  dig d1.dyn.jkl.mn @127.0.0.1 -p 5354
}


progname=$(basename $0)
subcommand=$1
case $subcommand in
    "" | "-h" | "--help")
        help
        ;;
    *)
        shift
        echo "Executing: $subcommand"
        ${subcommand} "$@"
        if [ $? = 127 ]; then
            echo "Error: '$subcommand' is not a known subcommand." >&2
            echo "       Run '$progname --help' for a list of known subcommands." >&2
            exit 1
        fi
        ;;
esac