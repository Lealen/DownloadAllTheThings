## Download All The Things!

This program lets you download all the files found on the page and follow them to download all files in other directories too. Tested and designed to work with default apache file list page, but may work with other similar too.

![Download All The Things](https://i.imgur.com/pwAzz8a.jpg)

### Setup

You should have a correctly installed Go compiler environment and your personal workspace ($GOPATH). If you have no idea what **$GOPATH** is, take a look [here](http://golang.org/doc/code.html). Please make sure that your **$GOPATH/bin** is available in your **$PATH**. 
Then you need to get the latest version of downloadallthethings, you can do this with this command:

    go get -u github.com/lealen/downloadallthethings

Done!
Run this command to see if everything is working:

    downloadallthethings --help

and then you can set the parameters like how many threads you want to use to download files and scrap pages. All arguments that are not a flags will be treated as an urls and tried to be downloaded.

Example:

    downloadallthethings -v -threads 4 [URL]
