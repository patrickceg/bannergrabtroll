# bannergrabtroll
Joke program for someone trying to poke at your TCP ports

## Overview

This program attempts to send pseudorandom garbage back at attackers who try to extract data from your transmission control protocol (TCP) ports.

### Status

It runs? I also discovered by running this program that there are banner grabbing attackers that will happily try to download gigabytes of pseudorandom garbage out of this app.

* Specifically the Telnet auto-attacking scripts seem to download the most junk, going up to gigabytes as mentioned
** For my personal runs I have since reduced the payload in case someone is using the tiny amount of bandwidth for DoS... but I don't personally know of a TCP based reflection DoS that you can do without messing with routers or other infrastructure between the bannergrabtroll app and the target
* The SSH and other random port banner grabbers seem to take a small amount of data and give up

## Disclaimers

This is just a hobby project for me rather than a proper tool for security. In fact, it can decrease your security:

* Of course you have ports that are open, so more attackers can detect your computer
* It's a drain of bandwidth to send the pseudorandom garbage
* This may be against a provider's terms of service (check your provider just in case): you'll probably trigger some "no servers" rules in particular
* If there's a problem in this program that is exploited, then you get hacked
* Do not modify this or do any variant of this for UDP!
*- An IP address in UDP is easily spoofed, so you can actually become part of a distributed denial of service (DDoS) attack if an attacker knows you're running something like this and gives you a packet with their intended target's address. For an example, read articles about _memcrashed_, the nickname of the DDoS against Github and others using the memcached database.

If you _actually_ want to get back at attackers, look up Honeypots, which are designed to log attacks and provide attackers something that looks like an actual server.

## Usage

### Basics

You run the compiled application, with at the very least the port to specify:

example that listens on port 4000

```
bannergrabapp -p 4000
```

The application will then ask you to acknowledge a similar disclaimer to the one above, which will require you to type in a value into the keyboard (or perhaps other app attached to STDIN). After accepting the disclaimer, the app will create a blank file in the directory you ran it from (so to not ask you again).

### Docker

If running with the included Dockerfile, you can build an image for it and run it. (I can't see a reason to make this joke pollute Docker hub at this point, although if you feel like it I won't care if someone forks this repository and points a Docker Hub build at it...?)

```
sudo docker build -t test .
sudo docker run --name bannergrabtrolltest -p 4000-4020:4000-4020 --rm test app -p 4000-4020
```

This will run the application on port 4000 to 4020.

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

* Throttle the bandwidth and resources used (done)
*- While many the banner grabbers disconnected (or crashed?) shortly after a few kilobytes of garbage, others sat around and got ~2 GB transferred over at almsot my full uplink, which isn't fun for ISP billing
* Allow multiple banner grabbers to get hit at once on the same port (done)
* Make more ports appear open (done)


