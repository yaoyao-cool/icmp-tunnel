

ICMP隧道工具，只是学习一下原理，代码有点丑

```azure
server:
    sudo ./icmp -t s
    
client:
    sudo ./icmp -l 7777 -sip 192.168.222.232 -tip 127.0.0.1 -tport 80
```

![img.png](img.png)
