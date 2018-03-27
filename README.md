# bannergrabtroll
Joke program for someone trying to poke at your TCP ports

## Overview

This program attempts to send pseudorandom garbage back at attackers who try to extract data from your transmission control protocol (TCP) ports.

### Status

I'll write more stuff as I get time

## Disclaimers

This is just a hobby project for me rather than a proper tool for security. In fact, it can decrease your security:

* Of course you have ports that are open, so more attackers can detect your computer
* It's a drain of bandwidth to send the pseudorandom garbage
* This may be against a provider's terms of service (check your provider just in case): you'll probably trigger some "no servers" rules in particular
* If there's a problem in this program that is exploited, then you get hacked
* Do not modify this or do any variant of this for UDP!
*- An IP address in UDP is easily spoofed, so you can actually become part of a distributed denial of service (DDoS) attack if an attacker knows you're running something like this and gives you a packet with their intended target's address. For an example, read articles about _memcrashed_, the nickname of the DDoS against Github and others using the memcached database.

If you _actually_ want to get back at attackers, look up Honeypots, which are designed to log attacks and provide attackers something that looks like an actual server.

## Inspiration

This is a little project I conceived of while watching my firewall and intrusion detection system (IDS) logs. Anything attached to the Internet gets poked and prodded by scripts that look for buffer overflows... so I thought - if those scripts are also vulnerable to buffer overflows, I can send some mischief back at the port scans / banner grabbing.

The start of this project was this little bit of shell script

```bash
#! /bin/bash

TRAP_IP=$1
while true; do
  echo Trap $TRAP_IP up
  dd if=/dev/urandom bs=16 count=134217728 | netcat -vl $TRAP_IP > /dev/null
  echo Trap $TRAP_IP triggered
  sleep 5
done
```

My goal is to learn go (I've never programmed in go before) and add these features to the joke script:

* Throttle the bandwidth and resources used
*- While many the banner grabbers disconnected (or crashed?) shortly after a few kilobytes of garbage, others sat around and got ~2 GB transferred over at almsot my full uplink, which isn't fun for ISP billing
* Allow multiple banner grabbers to get hit at once on the same port
* Make more ports appear open


