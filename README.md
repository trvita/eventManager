# eventManager

##gRPC client and server:

&nbspMakeEvent
&nbspGetEvent
&nbspDeleteevent
&nbspGetEvents

##AMQP broker:

&nbspExhangeDeclare
&nbspQueueDeclare
&nbspBind{ex, rk, queue}
&nbspPublish{ex, rk, event}
&nbspSubscribe{queue}  
rk - routing key = "senderID"  
ex - exchange = "senderID"  
queueName - "senderID + sessionID"

example: clients send to server commands 'MakeEvent'  
server sends to broker command Publish  
broker sends to clients their Events
