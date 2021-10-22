package kafkaques

type Flags struct {
	LogLevel string         `default:"info" enum:"error,warn,info,debug" help:"log level."`
	Produce  ProducerFlags `cmd:"" help:"Produce messages"`
	Consume  ConsumerFlags `cmd:"" help:"Consumer messages"`
}

type ConsumerFlags struct {
	Broker string   `kong:"required,help='Broker.'"`
	Group  string   `kong:"required,help='Group.'"`
	Topics []string `kong:"required,arg,name='topics',help='Topics to listen',type:'topics'"`
}

type ProducerFlags struct {
	Broker string `kong:"required,help='Broker.'"`
	Topic  string `kong:"required,arg,name='topic',help='Topic push messages to.',type:'topic'"`
}
