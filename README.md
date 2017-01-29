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