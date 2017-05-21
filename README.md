# Nexus


![Codeship Build Status](https://codeship.com/projects/789d4580-c880-0134-af42-6ac8e955f005/status?branch=master)


Nexus is a small websocket server framework.

If in json format, it sends and receives messages as shown in the tags

```
type Packet struct {
	Type string `json:"type"`
	Data string `json:"data"`
}
```


Using the delimited format serializes and deserializes faster by simply appending things together.

It sends and receives messages in the format:

```
# Format
<type_length>:<type><data>

# Examples
5:tokenTOKEN_STRING_GGGGGGGGGGGGGGGG
4:ping
4:chat{"message":"hello"}
```

Nexus provides a simple interface for registering a handler for each message type:

```golang

func main() {

	n := nexus.NewNexus()

	n.Handle("token", handleToken)
	n.Handle("chat", handleChat)

	http.ListenAndServe(":9090", n.Handler)
}

func handleToken(c *nexus.Client, p *nexus.Packet) {		
	user, err := lookupUserByToken(p.Data)
	if err != nil {
		// send a response here
		return
	}

	// store the user in the client's Env for using later
	c.Env["user"] = user
}

func handleChat(c *nexus.Client, p *nexus.Packet) {
	if _, ok := c.Env["user"] !ok {
		// unauthorized
		return
	}

	n.All().Broadcast(p)
}
```



Cheers! I'll add more later as I think of ways to keep it simple with websockets
