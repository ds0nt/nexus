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
type Hub struct {
	Hub *nexus.Nexus
}

func InitHub(db *storm.DB, mailerClient proto.MailerClient, containerClient container.ContainerClient) *Hub {
	h := &Hub{
		Hub:             nexus.NewNexus(),
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
	
	
func (h *Hub) handleToken(c *nexus.Client, p *nexus.Packet) {
	log.Println("handle token")
	in := tokenIn{}
	err := json.Unmarshal([]byte(p.Data), &in)
	if err != nil {
		log.Println("handleToken err unmarshalling", p.Data, err)
		return
	}
	c.Env["token"] = in.Token
	c.Env["authorized"] = jwt.Authorized(in.Token)
	user := ChatUser{
		Name: c.Env["authorized"].(string),
		Exec: make(chan string, 10),
	}
	c.Env["user"] = &user
	log.Println("authorized", user.Name)
	for u, _ := range allUsers {
		c.Send(buildJoinUserPacket(u))
	}
	allUsersMu.Lock()
	allUsers[&user] = true
	allUsersMu.Unlock()
	h.Hub.All().Broadcast(buildJoinUserPacket(&user))

	for u, ch := range channelsByName {
		c.Send(ircJoinPacket(ch.Nick, u))
	}
	chats := []savedChat{}
	err = h.db.All(&chats)
	if err != nil {
		log.Println("handleJoin err looking up saved chats", p.Data, err)
		return
	}

	data, _ := json.Marshal(&TTYResponse{"http://tty.ds0nt.com"})
	c.Send(&nexus.Packet{
		Type: "tty",
		Data: string(data),
	})

	l := int64(len(chats))
	i := l - 50
	if i < 0 {
		i = 0
	}
	for ; i < l; i++ {
		chat := chats[i]
		c.Send(ChatPacket(&ChatOut{
			From:     chat.From,
			Message:  chat.Message,
			Images:   chat.Images,
			Links:    chat.Links,
			Historic: true,
		}))

	}
}

```


Cheers! I'll add more later as I think of ways to keep it simple with websockets
