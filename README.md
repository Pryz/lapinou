# [WIP] Lapinou

```
   ***
  ** **
 **   **
 **   **         ****
 **   **       **   ****
 **  **       *   **   **
  **  *      *  **  ***  **
   **  *    *  **     **  *
    ** **  ** **        **
    **   **  **
   *           *
  *             *
 *    0     0    *
 *   /   @   \   *
 *   \__/ \__/   *
   *     W     *
     **     **
       *****
```

## What is Lapinou ?

Lapinou is a dumb tool to pin VCPU(s) of instances running with Qemu/KVM. The tool is leveraging [libvirt-go](https://github.com/Pryz/libvirt-go).

Right now Lapinou will try to pin all the VCPUs of all the Domains present on the host and will only pin per CPU Threads. We will maybe add more pinning strategies in the future.
See : [IBM - Tuning KVM for performance](ibm.com/support/knowledgecenter/linuxonibm/liaat/liaattuning_pdf.pdf).

*DISCLAIMER* : Pinning VCPU(s) can have really bad impact on your performance if you do know what you are doing. Do not use this tool in your prod :)

## Examples

### With only 1 Domain on the host

```
$ lapinou -cli
INFO[0000] Domains founds                                count=3
INFO[0000] Enough CPUs to apply pinning on provisioned VCPUs
INFO[0000] Working on domain                             name=instance-00000127
INFO[0000] Pinning VCPU on threads                       threads=16,40 vcpu=0
INFO[0000] Pinning VCPU on threads                       threads=17,41 vcpu=1
INFO[0000] Pinning VCPU on threads                       threads=18,42 vcpu=2
INFO[0000] Pinning VCPU on threads                       threads=19,43 vcpu=3
INFO[0000] Pinning VCPU on threads                       threads=20,44 vcpu=4
INFO[0000] Pinning VCPU on threads                       threads=21,45 vcpu=5
INFO[0000] Pinning VCPU on threads                       threads=22,46 vcpu=6
INFO[0000] Pinning VCPU on threads                       threads=23,47 vcpu=7
INFO[0000] Pinning VCPU on threads                       threads=0,24 vcpu=8
INFO[0000] Pinning VCPU on threads                       threads=1,25 vcpu=9
INFO[0000] Pinning VCPU on threads                       threads=2,26 vcpu=10
INFO[0000] Pinning VCPU on threads                       threads=3,27 vcpu=11
INFO[0000] Pinning VCPU on threads                       threads=4,28 vcpu=12
INFO[0000] Pinning VCPU on threads                       threads=5,29 vcpu=13
INFO[0000] Pinning VCPU on threads                       threads=6,30 vcpu=14
INFO[0000] Pinning VCPU on threads                       threads=7,31 vcpu=15
INFO[0000] Pinning VCPU on threads                       threads=8,32 vcpu=16
INFO[0000] Pinning VCPU on threads                       threads=9,33 vcpu=17
INFO[0000] Pinning VCPU on threads                       threads=10,34 vcpu=18
INFO[0000] Pinning VCPU on threads                       threads=11,35 vcpu=19
INFO[0000] Pinning VCPU on threads                       threads=12,36 vcpu=20
INFO[0000] Pinning VCPU on threads                       threads=13,37 vcpu=21
INFO[0000] Pinning VCPU on threads                       threads=14,38 vcpu=22
INFO[0000] Pinning VCPU on threads                       threads=15,39 vcpu=23
INFO[0000] Pinning VCPU on threads                       threads=16,40 vcpu=24
INFO[0000] Pinning VCPU on threads                       threads=17,41 vcpu=25
INFO[0000] Pinning VCPU on threads                       threads=18,42 vcpu=26
INFO[0000] Pinning VCPU on threads                       threads=19,43 vcpu=27
INFO[0000] Pinning VCPU on threads                       threads=20,44 vcpu=28
INFO[0000] Pinning VCPU on threads                       threads=21,45 vcpu=29
INFO[0000] Pinning VCPU on threads                       threads=22,46 vcpu=30
INFO[0000] Pinning VCPU on threads                       threads=23,47 vcpu=31
```

### With multiple Domains on the host

```
$ lapinou -cli
INFO[0000] Domains founds                                count=3
INFO[0000] Enough CPUs to apply pinning on provisioned VCPUs
INFO[0000] Working on domain                             name=instance-00000006
INFO[0000] Pinning VCPU on threads                       cpu=16 threads=4,16 vcpu=0
INFO[0000] Pinning VCPU on threads                       cpu=17 threads=5,17 vcpu=1
INFO[0000] Pinning VCPU on threads                       cpu=18 threads=6,18 vcpu=2
INFO[0000] Pinning VCPU on threads                       cpu=19 threads=7,19 vcpu=3
INFO[0000] Pinning VCPU on threads                       cpu=20 threads=8,20 vcpu=4
INFO[0000] Pinning VCPU on threads                       cpu=21 threads=9,21 vcpu=5
INFO[0000] Pinning VCPU on threads                       cpu=22 threads=10,22 vcpu=6
INFO[0000] Pinning VCPU on threads                       cpu=23 threads=11,23 vcpu=7
INFO[0000] Working on domain                             name=instance-00000004
INFO[0000] Pinning VCPU on threads                       cpu=8 threads=8,20 vcpu=0
INFO[0000] Pinning VCPU on threads                       cpu=9 threads=9,21 vcpu=1
INFO[0000] Pinning VCPU on threads                       cpu=10 threads=10,22 vcpu=2
INFO[0000] Pinning VCPU on threads                       cpu=11 threads=11,23 vcpu=3
INFO[0000] Pinning VCPU on threads                       cpu=12 threads=0,12 vcpu=4
INFO[0000] Pinning VCPU on threads                       cpu=13 threads=1,13 vcpu=5
INFO[0000] Pinning VCPU on threads                       cpu=14 threads=2,14 vcpu=6
INFO[0000] Pinning VCPU on threads                       cpu=15 threads=3,15 vcpu=7
INFO[0000] Working on domain                             name=instance-00000003
INFO[0000] Pinning VCPU on threads                       cpu=4 threads=4,16 vcpu=0
INFO[0000] Pinning VCPU on threads                       cpu=5 threads=5,17 vcpu=1
INFO[0000] Pinning VCPU on threads                       cpu=6 threads=6,18 vcpu=2
INFO[0000] Pinning VCPU on threads                       cpu=7 threads=7,19 vcpu=3
```

## Run it

Use it manually or inside a crontab. The work to daemonize the thing is in progress.

```
*/30 * * * * /usr/local/bin/lapinou -cli > /tmp/lapinou.log
```
If you want JSON logs

```
*/30 * * * * /usr/local/bin/lapinou -cli -jsonlog > /tmp/lapinou.log
```


## Build and deploy Lapinou

To build Lapinou you will need a box/container with libvirt-dev installed, go and make. Then to build the binary :

```
make
```

To build the package (You will need FPM for that) :

```
make deb
```
