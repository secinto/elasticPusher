<h1 align="center">elasticPusher</h1>
<h4 align="center">Client for pushing data to ELK</h4>
<p align="center">
  
  <img src="https://img.shields.io/github/watchers/secinto/elasticPusher?label=Watchers&style=for-the-badge" alt="GitHub Watchers">
  <img src="https://img.shields.io/github/stars/secinto/elasticPusher?style=for-the-badge" alt="GitHub Stars">
  <img src="https://img.shields.io/github/license/secinto/elasticPusher?style=for-the-badge" alt="GitHub License">
</p>

Developed by Stefan Kraxberger (https://twitter.com/skraxberger/)  

Released as open source by secinto GmbH - https://secinto.com/  
Released under Apache License version 2.0 see LICENSE for more information

Description
----
elasticPusher is a GO client tool which pushes specified files to logstash. Different file types can be processed. 
Currently JSONL and RESPONSE and RAW. RAW can currently be used for any file type which is not JSON based. It is wrapped into an 
"interaction" JSON container. RESPONSE is used for the response output saved by HTTPX, this is parsed into a more useful 
structure, for better readability.

# Installation Instructions

`elasticPusher` requires **go1.20** to install successfully. Run the following command to get the repo:

```sh
git clone https://github.com/secinto/elasticPusher.git
cd elasticPusher
go build
go install
```

or the following to directly install it from the command line:

```sh
go install -v github.com/secinto/elasticPusher/cmd/elasticPusher@latest
```

# Usage

```sh
elasticPusher -help
```

This will display help for the tool. Here are all the switches it supports.


```console
Usage:
  elasticPusher [flags]

Flags:
INPUT:
   -f, -file string     input file containing data to be stored
   -i, -index string    index under which the data should be stored
   -t, -type string     input is in JSONL(ines) or raw (HTTPX response output) format (default "json")
   -p, -project string  project name for metadata addition
   -h, -host string     host name for metadata addition

CONFIG:
   -config string  flag configuration file (default "$HOME/.config/elasticPusher/config.yaml")

DEBUG:
   -silent         show only results in output
   -version        show version of the project
   -v              show verbose output
   -nc, -no-color  disable colors in output


