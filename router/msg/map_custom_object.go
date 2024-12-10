package msg

import fmt "fmt"

func (b *MsgBuilder) WithReqCustomObject(v any) *MsgBuilder {

	byt, err := b.wrapper.parm.CustomObjSerialize.Encode(v)
	if err != nil {
		b.wrapper.Err = fmt.Errorf("with req custom object encode err %v", err)
	}

	b.wrapper.Req.Header.Custom = byt

	return b
}

func (b *MsgBuilder) WithResCustomObject(v any) *MsgBuilder {
	byt, err := b.wrapper.parm.CustomObjSerialize.Encode(v)
	if err != nil {
		b.wrapper.Err = fmt.Errorf("with res custom object encode err %v", err)
	}

	b.wrapper.Res.Header.Custom = byt

	return b
}

func (b *MsgBuilder) GetReqCustomObject(v any) error {
	data := b.wrapper.Req.Header.Custom
	if len(data) == 0 {
		return fmt.Errorf("get req custom object is empty")
	}

	return b.wrapper.parm.CustomObjSerialize.Decode(data, v)
}

func (b *MsgBuilder) GetResCustomObject(v any) error {
	data := b.wrapper.Res.Header.Custom
	if len(data) == 0 {
		return fmt.Errorf("get res custom object is empty")
	}

	return b.wrapper.parm.CustomObjSerialize.Decode(data, v)
}
