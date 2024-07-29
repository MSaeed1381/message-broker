package store

type Config struct {
	inMemory bool
}

func StorageConfig() Config {
	return Config{
		inMemory: true,
	}
}

//
//func NewTopic() Topic {
//	if StorageConfig().inMemory {
//		return memory.NewTopicInMemory()
//	} else {
//		return nil
//	}
//}
//
//func NewMessage() Message {
//	if StorageConfig().inMemory {
//		return memory.NewMessageInMemory()
//	} else {
//		return nil
//	}
//}
