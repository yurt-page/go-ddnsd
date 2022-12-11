# go-ddnsd
Dynamic DNS server with DynDNS API support and automatic registration

## Run the ddnsd locally
To start locally for debugging you need to specify ports and IPs like this:

    ddnsd --listen-dns=:5354 --listen-api=:8080 --self-domain=dyn.jkl.mn. --self-ip=127.0.0.1 --verbose

## Configure OpenWrt DDNS
Install `ddns-scripts` and optionally `luci-app-ddns` and `ddns-scripts-services`:

    opkg install ddns-scripts ddns-scripts-services luci-app-ddns

Add this to `/etc/config/ddns`:

```
config service 'jklmn'
    option enabled '1'
    option update_url 'http://[USERNAME]:[PASSWORD]@dyn.jkl.mn:8080/nic/update?hostname=[DOMAIN]&myip=[IP]'
    option lookup_host 'YOURDOMAIN.dyn.jkl.mn.'
    option domain 'YOURDOMAIN.dyn.jkl.mn.'
    option username 'nologin'
    option password 'VERY_VERY_LONG_PASSWORD'
    option ip_source 'network'
	option ip_network 'wan'
	option interface 'wan'
```

Then restart `service ddns restart`

You may use UI at http://192.168.1.1/cgi-bin/luci/admin/services/ddns

## DynDNS API description
All routers already have a support of Dyn.com (previously called DynDNS.com).
A router just makes a GET request `/nic/update?hostname={YOURDOMAIN}` to a server.
It detects an IP and updates a DNS record.
Almost all DynDNS providers supports the same API endpoint as Dyn.com.
The API is unofficially called DynDNS2 i.e. DynDNS.com version 2.

Some protocol descriptions:
* https://help.dyn.com/remote-access-api/perform-update/
* https://support.google.com/domains/answer/6147083
* https://www.dynu.com/en-US/DynamicDNS/IP-Update-Protocol


## License
[0BSD](https://opensource.org/licenses/0BSD) (similar to Public Domain)