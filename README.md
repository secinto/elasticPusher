<h1 align="center">elasticPusher</h1>
<h4 align="center">Advanced Logging for Burp Suite</h4>
<p align="center">
  
  <img src="https://img.shields.io/github/watchers/secinto/elasticPusher?label=Watchers&style=for-the-badge" alt="GitHub Watchers">
  <img src="https://img.shields.io/github/stars/secinto/elasticPusher?style=for-the-badge" alt="GitHub Stars">
  <img src="https://img.shields.io/github/downloadssecinto/elasticPusher/total?style=for-the-badge" alt="GitHub All Releases">
  <img src="https://img.shields.io/github/license/secinto/elasticPusher?style=for-the-badge" alt="GitHub License">
</p>

Developed by Stefan Kraxberger (https://twitter.com/skraxberger/)  

Released as open source by secinto GmbH - https://secinto.com/  
Released under Apache License version 2.0 see LICENSE for more information

Description
----
elasticPusher is a GO client tool which pushes specified files to logstash. Different file types can be processed. 
Currently JSONL and RAW. RAW can currently be used for any file type which is not JSON based. It is wrapped into an 
"interaction" JSON container, mostly used for storing HTTP request/responses. 

# Installation Instructions

`elasticPusher` requires **go1.19** to install successfully. Run the following command to get the repo:

```sh
git clone https://github.com/secinto/elasticPusher.git
cd elasticPusher
go build
go install
```

# Usage

```sh
elasticPusher -help
```

This will display help for the tool. Here are all the switches it supports.


```console
Usage:
  ./elasticPusher [flags]

Flags:
   -f,                   input file containing the data to be stored in elastic/logstash
   -i                    the index under which the content should be stored
   -p                    project name which will be added as additional information to the data
   -h                    host name which will be added as additional information to the data
