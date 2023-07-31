package hook

import (
	"fmt"
	"strings"
)

func SplitTopicACL(fullTopic string) (topic, acl string, err error) {
	// topic/private
	topicExtracted := strings.Split(fullTopic, "/")
	if len(topicExtracted) == 1 {
		if len(topicExtracted[0]) == 0 {
			err = fmt.Errorf("Invalid full topic: %s", fullTopic)
			return
		}
		return topicExtracted[0], "", nil
	}
	if len(topicExtracted) == 2 {
		if len(topicExtracted[0]) == 0 {
			err = fmt.Errorf("Invalid full topic: %s", fullTopic)
			return
		}
		if len(topicExtracted[1]) == 0 {
			err = fmt.Errorf("Invalid full topic: %s", fullTopic)
			return
		}
		return topicExtracted[0], topicExtracted[1], nil
	}
	err = fmt.Errorf("Invalid full topic: %s", fullTopic)
	return
}

func SplitTopicUserNameFromFullUsername(
	fullUserName string,
) (userName, topic string, err error) {
	userNameExtracted := strings.Split(fullUserName, "/")
	if len(userNameExtracted) != 2 {
		err = fmt.Errorf("Invalid full username: %s", fullUserName)
		return
	}
	if len(userNameExtracted[0]) == 0 {
		err = fmt.Errorf("Invalid username: %s", userNameExtracted[0])
		return
	}
	if len(userNameExtracted[1]) == 0 {
		err = fmt.Errorf("Invalid topic: %s", userNameExtracted[1])
		return
	}
	return userNameExtracted[0], userNameExtracted[1], nil
}
