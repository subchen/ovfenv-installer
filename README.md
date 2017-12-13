# ovfenv-installer

[![Build Status](https://travis-ci.org/subchen/frep.svg?branch=master)](https://travis-ci.org/subchen/frep)
[![License](http://img.shields.io/badge/License-Apache_2-red.svg?style=flat)](http://www.apache.org/licenses/LICENSE-2.0)

Configure networking from vSphere ovfEnv properties (CentOS 7.0+ Only)

```
NAME:
   ovfenv-installer - Configure networking from vSphere ovfEnv properties

USAGE:
   ovfenv-installer [ OPTIONS ]

VERSION:
   1.0.1-10

AUTHORS:
   Guoqiang Chen <subchen@gmail.com>

OPTIONS:
   --run-once        run once only
   --log-file path   save log to file
   --help            print this usage
   --version         print version information

EXAMPLES:
   You can append following command-line into /etc/rc.d/rc.local (chmod +x)
   >> ovfenv-installer --run-once --log-file=/var/log/ovfenv-installer.log
```

## Downloads

v1.0.1 Released: 
https://github.com/subchen/ovfenv-installer/releases/tag/v1.0.1

## Examples:

See https://github.com/subchen/centos-7-kickstart

