# Nexus


![Codeship Build Status](https://codeship.com/projects/789d4580-c880-0134-af42-6ac8e955f005/status?branch=master)


Nexus is a websocket framework that takes care of common websocket tasks, and creates simple objects for use in web applications.

 - Clients are created for every websocket connection.
 - Handlers deal with packets sent by clients, and are routed based on the "type" json key
 - Clients have a client.Send(p *Packet) method
 - Pools are groups of clients.
   - Pools have a pool.Broadcast(p *Packet) and a pool.Add(c *Client) 
   - Clients are automatically removed from all the pools when they disconnect.

Nexus also saves you from spending your whole day un-deadlocking your websocket code. <3

```golang
unc InitHub(db *storm.DB, mailerClient proto.MailerClient, containerClient container.ContainerClient) *Hub {
	h := &Hub{
		Hub:             nexus.NewNexus(),
		OnlineCount:     0,
		db:              db,
		mailerClient:    mailerClient,
		containerClient: containerClient,
	}
	db.Init(&savedChat{})

 h.Hub.Handle("token", h.handleToken)
	h.Hub.Handle("ping", h.handlePing)
	h.Hub.Handle("join", h.handleJoin)
	h.Hub.Handle("chat", h.handleChat)

	h.Hub.All().AfterAdd(func(p *nexus.Pool, c *nexus.Client) {
		atomic.AddInt64(&h.OnlineCount, 1)
		h.Hub.All().Broadcast(&nexus.Packet{
			Type: "online",
			Data: strconv.Itoa(int(h.OnlineCount)),
		})

	})
	h.Hub.All().AfterRemove(func(p *nexus.Pool, c *nexus.Client) {
		atomic.AddInt64(&h.OnlineCount, -1)
		h.Hub.All().Broadcast(&nexus.Packet{
			Type: "online",
			Data: strconv.Itoa(int(h.OnlineCount)),
		})

		if user, ok := c.Env["user"]; ok {
			cUser := user.(*ChatUser)
			allUsersMu.Lock()
			delete(allUsers, cUser)
			allUsersMu.Unlock()
			p.Broadcast(&nexus.Packet{
				Type: "leave-user",
				Data: cUser.Name,
			})

		}
	})
```


Cheers! I'll add more later as I think of ways to keep it simple with websockets
