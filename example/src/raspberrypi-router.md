---
title: Raspberry Pi as a router
date: 2016-03-29
---

One possible way to setup a Raspberry Pi as a router is to flash [OpenWRT](https://wiki.openwrt.org/toh/raspberry_pi_foundation/raspberry_pi) on an SD card.
I will try that when I have an extra SD card.
But for now I wanted to have it work as a router and keep using Raspbian at the same time.

This method is quite simple. It invloves adding iptable rules and optionally using sshuttle to make it work as a proxy.
All devices connected to this wifi network will have their traffic proxied through the Pi.
Here is the overall setup of my LAN
(the powerline extender is optional indeed - I have to use it due to the topology of our house):

ISP modem &hArr; Raspberry Pi &hArr; Powerline extender &hArr; Powerline Wifi

I connected eth0 to the ISP modem, and eth1 (via a USB-Ethernet adapter) to the powerline extender.
Here is a script to setup traffic forwarding in Raspberry Pi:


    # eth0 is connected to ISP modem
    # eth1 is connected to LAN extender
    
    # This should work, otherwise try editing /etc/sysctl.conf
    echo 1 > /proc/sys/net/ipv4/ip_forward
    sysctl -p
    
    # Always accept loopback traffic
    iptables -A INPUT -i lo -j ACCEPT
    
    # We allow traffic from the LAN side
    iptables -A INPUT -i eth0 -j ACCEPT
    
    # Allow established connections
    iptables -A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT
    
    # Masquerade.
    iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE
    
    # fowarding
    iptables -A FORWARD -i eth0 -o eth1 \
             -m state --state RELATED,ESTABLISHED -j ACCEPT
    
    # Allow outgoing connections from the LAN side.
    iptables -A FORWARD -i eth1 -o eth0 -j ACCEPT


Optionally, start sshuttle like this:

```shell
sshuttle --dns -vr host \
    -l 0.0.0.0          \
    -x 192.168.0.0/16   \
    0/0
```


I exclude 192.168.0.0/16, so I could still SSH into Pi without sshuttle proxying my connection to the host.

It is very difficult to find a Raspberry Pi where I live.
I would like to thank my friends
[@gluegadget](https://twitter.com/gluegadget),
[@cubny](https://twitter.com/cubny) and
[@artlesshand](https://twitter.com/artlesshand)
who got me this Raspberry Pi.
