package ndn

import (
	"errors"
	"github.com/davecgh/go-spew/spew"
	"time"
)

/*
   interact with NFD
*/

const (
	CONTROL_PARAMETERS    uint64 = 104
	FACE_ID                      = 105
	URI                          = 114
	LOCAL_CONTROL_FEATURE        = 110
	COST                         = 106
	STRATEGY                     = 107
	CONTROL_RESPONSE             = 101
	STATUS_CODE                  = 102
	STATUS_TEXT                  = 103
)

var (
	controlResponseFormat = node{Type: CONTROL_RESPONSE, Children: []node{
		{Type: STATUS_CODE},
		{Type: STATUS_TEXT},
		{Type: CONTROL_PARAMETERS, Count: ZERO_OR_ONE, Children: controlParametersContentFormat},
	}}
	controlParametersContentFormat = []node{
		{Type: NAME, Count: ZERO_OR_ONE, Children: []node{{Type: NAME_COMPONENT, Count: ZERO_OR_MORE}}},
		{Type: FACE_ID, Count: ZERO_OR_ONE},
		{Type: URI, Count: ZERO_OR_ONE},
		{Type: LOCAL_CONTROL_FEATURE, Count: ZERO_OR_ONE},
		{Type: COST, Count: ZERO_OR_ONE},
		{Type: STRATEGY, Count: ZERO_OR_ONE, Children: []node{
			nameFormat,
		}},
	}
	controlFormat = node{Type: NAME, Children: []node{
		{Type: NAME_COMPONENT}, // localhost
		{Type: NAME_COMPONENT}, // nfd
		{Type: NAME_COMPONENT}, // module
		{Type: NAME_COMPONENT}, // command
		{Type: NAME_COMPONENT, Children: []node{
			{Type: CONTROL_PARAMETERS, Children: controlParametersContentFormat}, // param
		}},
		{Type: NAME_COMPONENT}, // timestamp
		{Type: NAME_COMPONENT}, // random value
		{Type: NAME_COMPONENT, Children: []node{signatureInfoFormat}},
		{Type: NAME_COMPONENT, Children: []node{
			{Type: SIGNATURE_VALUE},
		}},
	}}
)

type Control struct {
	Module     string
	Command    string
	Parameters Parameters
}

type Parameters struct {
	Name                [][]byte
	FaceId              uint64
	Uri                 string
	LocalControlFeature uint64
	Cost                uint64
	Strategy            [][]byte
}

func (this *Parameters) Encode() (parameters TLV, err error) {
	parameters = NewTLV(CONTROL_PARAMETERS)
	// name
	if len(this.Name) != 0 {
		parameters.Add(nameEncode(this.Name))
	}
	// face id
	if this.FaceId != 0 {
		faceId := NewTLV(FACE_ID)
		faceId.Value, err = encodeNonNeg(this.FaceId)
		if err != nil {
			return
		}
		parameters.Add(faceId)
	}
	// uri
	if len(this.Uri) != 0 {
		uri := NewTLV(URI)
		uri.Value = []byte(this.Uri)
		parameters.Add(uri)
	}
	// local control feature
	if this.LocalControlFeature != 0 {
		localControlFeature := NewTLV(LOCAL_CONTROL_FEATURE)
		localControlFeature.Value, err = encodeNonNeg(this.LocalControlFeature)
		if err != nil {
			return
		}
		parameters.Add(localControlFeature)
	}
	// cost
	if this.Cost != 0 {
		cost := NewTLV(COST)
		cost.Value, err = encodeNonNeg(this.Cost)
		if err != nil {
			return
		}
		parameters.Add(cost)
	}
	// strategy
	if len(this.Strategy) != 0 {
		strategy := NewTLV(STRATEGY)
		strategy.Add(nameEncode(this.Strategy))
		parameters.Add(strategy)
	}
	return
}

func (this *Parameters) Decode(parameters TLV) (err error) {
	for _, c := range parameters.Children {
		switch c.Type {
		case NAME:
			this.Name = nameDecode(c)
		case FACE_ID:
			this.FaceId, err = decodeNonNeg(c.Value)
			if err != nil {
				return
			}
		case URI:
			this.Uri = string(c.Value)
		case LOCAL_CONTROL_FEATURE:
			this.LocalControlFeature, err = decodeNonNeg(c.Value)
			if err != nil {
				return
			}
		case COST:
			this.Cost, err = decodeNonNeg(c.Value)
			if err != nil {
				return
			}
		case STRATEGY:
			this.Strategy = nameDecode(c.Children[0])
		}
	}
	return
}

func (this *Control) Print() {
	spew.Dump(*this)
}

func (this *Control) Encode() (i *Interest, err error) {
	name := nameFromString("/localhost/nfd/" + this.Module + "/" + this.Command)
	parameters, err := this.Parameters.Encode()
	if err != nil {
		return
	}
	b, err := parameters.Encode()
	if err != nil {
		return
	}
	name = append(name, b)
	// signed

	// timestamp
	b, err = encodeNonNeg(uint64(time.Now().UnixNano() / 1000000))
	if err != nil {
		return
	}
	name = append(name, b)
	// random value
	name = append(name, newNonce())

	// signature info
	signatureInfo := NewTLV(SIGNATURE_INFO)
	// signature type
	signatureType := NewTLV(SIGNATURE_TYPE)
	signatureType.Value, err = encodeNonNeg(SIGNATURE_TYPE_SIGNATURE_SHA_256_WITH_RSA)
	if err != nil {
		return
	}
	signatureInfo.Add(signatureType)
	// add empty keyLocator for rsa
	keyLocator := NewTLV(KEY_LOCATOR)
	keyLocator.Add(nameEncode(nameFromString("/testing/KEY/pubkey/ID-CERT")))
	signatureInfo.Add(keyLocator)

	b, err = signatureInfo.Encode()
	if err != nil {
		return
	}
	name = append(name, b)

	// signature value
	signatureValue := NewTLV(SIGNATURE_VALUE)
	signatureValue.Value, err = signRSA(nameEncode(name).Children)
	if err != nil {
		return
	}
	b, err = signatureValue.Encode()
	if err != nil {
		return
	}
	name = append(name, b)

	// final encode
	i = NewInterest("")
	i.Name = name
	i.Selectors.MustBeFresh = true
	return
}

func DecodeControl(name []byte) (ctrl TLV, err error) {
	ctrl, remain, err := matchNode(controlFormat, name)
	if err != nil {
		return
	}
	if len(remain) != 0 {
		err = errors.New("buffer not empty")
	}
	return
}

func (this *Control) Decode(i *Interest) (err error) {
	name := nameEncode(i.Name)
	b, err := name.Encode()
	if err != nil {
		return
	}
	ctrl, err := DecodeControl(b)
	if err != nil {
		return
	}
	// module
	this.Module = string(ctrl.Children[2].Value)
	// command
	this.Command = string(ctrl.Children[3].Value)
	// parameters
	err = this.Parameters.Decode(ctrl.Children[4].Children[0])
	if err != nil {
		return
	}

	// TODO: enable rsa
	// signatureValue := ctrl.Children[8].Children[0].Value
	// if !verifyRSA(ctrl.Children[:8], signatureValue) {
	// 	err = errors.New("cannot verify rsa")
	// 	return
	// }
	return
}

type ControlResponse struct {
	StatusCode uint64
	StatusText string
	Body       Parameters
}

const (
	STATUS_CODE_OK             uint64 = 200
	STATUS_CODE_ARGS_INCORRECT        = 400
	STATUS_CODE_NOT_AUTHORIZED        = 403
	STATUS_CODE_NOT_FOUND             = 404
	STATUS_CODE_NOT_SUPPORTED         = 501
)

func DecodeControlResponse(content []byte) (resp TLV, err error) {
	resp, remain, err := matchNode(controlResponseFormat, content)
	if err != nil {
		return
	}
	if len(remain) != 0 {
		err = errors.New("buffer not empty")
	}
	return
}

func (this *ControlResponse) Print() {
	spew.Dump(*this)
}

func (this *ControlResponse) Encode() (d *Data, err error) {
	controlResponse := NewTLV(CONTROL_RESPONSE)
	// status code
	statusCode := NewTLV(STATUS_CODE)
	statusCode.Value, err = encodeNonNeg(this.StatusCode)
	controlResponse.Add(statusCode)
	// status text
	statusText := NewTLV(STATUS_TEXT)
	statusText.Value = []byte(this.StatusText)
	controlResponse.Add(statusText)

	// parameters
	parameters, err := this.Body.Encode()
	if err != nil {
		return
	}
	if len(parameters.Children) != 0 {
		controlResponse.Add(parameters)
	}

	d = &Data{}
	d.Content, err = controlResponse.Encode()
	return
}

func (this *ControlResponse) Decode(d *Data) error {
	resp, err := DecodeControlResponse(d.Content)
	if err != nil {
		return err
	}
	for _, c := range resp.Children {
		switch c.Type {
		case STATUS_CODE:
			this.StatusCode, err = decodeNonNeg(c.Value)
			if err != nil {
				return err
			}
		case STATUS_TEXT:
			this.StatusText = string(c.Value)
		case CONTROL_PARAMETERS:
			err = this.Body.Decode(c)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
