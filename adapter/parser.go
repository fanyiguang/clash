package adapter

//import (
//	C "github.com/Dreamacro/clash/constant"
//)

//func ParseInbound(mapping map[string]any, tcpIn chan<- C.ConnContext, udpIn chan<- *inbound.PacketAdapter) (C.OtherInbound, error) {
//	decoder := structure.NewDecoder(structure.Option{TagName: "json", WeaklyTypedInput: true})
//	otherInbound, existType := mapping["type"].(string)
//	if !existType {
//		return nil, fmt.Errorf("missing type")
//	}
//	var (
//		inbound C.OtherInbound
//		err     error
//	)
//	switch otherInbound {
//	case "socks":
//		socksOption := &otherinbound.SocksOption{}
//		err = decoder.Decode(mapping, socksOption)
//		if err != nil {
//			break
//		}
//		inbound, err = otherinbound.NewSocks(*socksOption, tcpIn, udpIn)
//	case "http":
//		httpOption := &otherinbound.HttpOption{}
//		err = decoder.Decode(mapping, httpOption)
//		if err != nil {
//			break
//		}
//		inbound, err = otherinbound.NewHttp(*httpOption, tcpIn)
//	case "direct":
//		directOption := &otherinbound.DirectOption{}
//		err = decoder.Decode(mapping, directOption)
//		if err != nil {
//			break
//		}
//		inbound, err = otherinbound.NewDirect(*directOption, tcpIn, udpIn)
//	default:
//		return nil, fmt.Errorf("unsupported proxy type: %s", otherInbound)
//	}
//	return inbound, err
//}
