# CoreDNS - ads Plugin

[![Build Status](https://cloud.drone.io/api/badges/c-mueller/ads/status.svg)](https://cloud.drone.io/c-mueller/ads)
[![codecov](https://codecov.io/gh/c-mueller/ads/branch/master/graph/badge.svg)](https://codecov.io/gh/c-mueller/ads)

DNS AdBlocker plugin for CoreDNS.

## Compiling

First get the CoreDNS source code by running, after you cloned this repository into the proper path in your `GOPATH`
```bash
go get github.com/coredns/coredns
```

Then navigate to the coredns directory
```bash
cd $(go env GOPATH)/src/github.com/coredns/coredns
```

Next update the `plugin.cfg` in the root of the coredns repository as follows

```bash
sed -i 's|loadbalance:loadbalance|ads:github.com/c-mueller/ads\nloadbalance:loadbalance|g' plugin.cfg
```

while I would suggest having the `ads` plugin before the `cache` plugin because it will
cause the changes in the blocklists to be applied instantly. However the overall performance of
the DNS server could degrade when having many regex rules. In that case I recommend putting the
plugin before the `hosts` plugin:

```bash
sed -i 's|hosts:hosts|ads:github.com/c-mueller/ads\nhosts:hosts|g' plugin.cfg
```

Finally run `make` to build CoreDNS with the `ads` plugin

The releases section also contains binaries of the latest CoreDNS with the
ads plugin. These get built automatically using drone. Once they have been triggered.

### Building development binaries

First build the binaries once using the process described above to ensure the `go.mod` file is fully generated.

Assuming the ads plugin is pulled in the appropriate path run the following command in the root of the CoreDNS binary:
```
For BASH:
echo "replace github.com/c-mueller/ads "$(cat go.mod | grep 'github.com/c-mueller/ads ' | sed 's|github.com/c-mueller/ads ||g; s|\t||g')" => "$(go env GOPATH)"/src/c-mueller/ads" >> go.mod

For Fish:
echo "replace github.com/c-muéller/ads "(cat go.mod | grep 'github.com/c-mueller/ads ' | sed 's|github.com/c-mueller/ads ||g; s|\t||g')" => "(go env GOPATH)"/src/c-mueller/ads" >> go.mod
```

After that just run `make` now CoreDNS should have complied with the version of the `ads` plugin you have currently checked out in the ads repository.

## Configuring

### Default settings

Running the `ads` plugin with all defaults is done by just adding the `ads` keyword to your Corefile.

For example:
```
.:53 {
    ads
    forward . 1.1.1.1
    log
    errors
}
```

### Configuring the `ads` plugin

You can see a more complex `ads` configuration in the following Corefile

```
.:53 {
   ads {
        list http://url-to-my-blocklists.com/list1.txt
        list http://url-to-my-blocklists.com/list2.txt
        default-lists
        blacklist google.com
        whitelist googleadservices.com
        target 10.133.7.8
        target-ipv6 ::1
   }
   # Other directives have been omitted
}
```

#### Configuration options

- `list <LIST URL>` HTTP(S)-URL to a hostlist to Block
- `default-lists` Readds the default hostlists to the internal list of blocklists.
    - This command is needed if you want to add custom blocklists and you want to also use the default ones
- `target <IPv4 IP>` defines the target ip to which blocked domains should resolve to if a A record is requested
- `target-ipv6 <IPv6 IP>` defines the target IPv6 address to which blocked domains should resolve to if a AAAA record is requested
- `disable-auto-update` Turns off the automatic update of the blocklists every 24h (can be changed)
- `log` Print a message every time a request gets blocked
- `auto-update-interval <INTERVAL>` Allows the modification of the interval between blocklist updates
    - This operation uses Golangs `time.ParseDuration()` function in order to parse the duration.
    Please ensure the specified duration can be parsed by this operation. Please refer to [here](https://golang.org/pkg/time/#ParseDuration).
    - This gets ignored if the automatic blocklist updates have been disabled
- `blocklist-path <FILEPATH FOR PERSISTED BLOCKLIST>` This option enables persisting of the blocklist
  to prevent a automatic redownload everytime CoreDNS restarts. The lists get persisted everytime a update get performed.
    - If autoupdates have been turned off the list will be reloaded every time the application launches.
    Making this option pretty useless for this kind of configuration.
- `whitelist <QNAME>` and `blacklist <QNAME>` Allows the explicit whitelisting or blacklisting of specific qnames. If a qname is on the whitelist it will not be blocked. 
- `whitelist-regex <REGEX>` and `blacklist-regex <REGEX>` identical to the regular whitelist and blacklist options. But instead of blocking a specific qname blocking is done for a regular expression. Yo might want to define exceptions to a regex blacklist entry. This can be done by using eitehr the `whitelist` or `whitelist-regex` options. 
