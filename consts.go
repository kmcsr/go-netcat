
package main

const LICENSE = 
`go-netcat  Copyright (C) 2023  Kevin Z <zyxkad@gmail.com>
This program comes with ABSOLUTELY NO WARRANTY.
This is free software, and you are welcome to redistribute it under certain conditions.
`

const helpMessage = `
Examples:
  go-netcat google.com:80
    Open a tcp connection in text mode to google.com:80
  go-netcat -u -l :12345 127.0.0.1:12345
    Open a udp connection in text mode at [::]:12345, and then connect to it (let's chat with yourself)
  go-netcat -u -b 8.8.8.8:53
    Open a udp connection to a DNS server with binary mode
`
