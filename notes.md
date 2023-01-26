Jan 25

Watch and follow first video. To spice it up, i implemented a unix domain
socket transport rather than tcp. Maybe a slow idea, because i ended up
needing to sort out how to differentiate incomming connections.

Initially the were anonymous, which was a by product of using net.Dial

instead, need net.DialUnix with an explicit filename for the local addr

this was helpful info
https://stackoverflow.com/questions/9644251/how-do-unix-domain-sockets-differentiate-between-multiple-clients
