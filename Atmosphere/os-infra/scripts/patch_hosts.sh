#!/bin/bash
printf "10.10.10.11 ctl1 ctl1.${domain}\n10.10.10.12 ctl2 ctl2.${domain}\n10.10.10.13 ctl3 ctl3.${domain}\n10.10.10.101 k8s.${domain}\n" >> /etc/hosts