
# go netcat

A netcat cli tool written in golang

### Install

```sh
go install github.com/kmcsr/go-netcat@latest
```

### Text Mode vs. Binary Mode

- Text mode will process data by lines, and it won't escape/unescape chars
   - If you want to specific the linebreak to be send, use flag `-linebreak`, avaliable values are:  
      `cr`: use `\r`  
      `lf`: use `\n` (default as well as fallback)  
      `crlf`: use `\r\n`
- Binary mode will process input data by lines with space splited numbers as bytes, and output data by per buffer.
   - The length of the buffer can be set use flag `-buffer` / `-B`
   - If the data cannot fill over the buffer, it will be print anyway
   - The base prefix is allowed, and the default base is 16 (`0b`: 2, `0o`: 8, `0x`: 16)
   - If you want to send literary text under the binray mode, just add prefix char `!` in front of the line
