// Copyright 2020, Chef.  All rights reserved.
// https://github.com/q191201771/lal
//
// Use of this source code is governed by a MIT-style license
// that can be found in the License file.
//
// Author: Chef (191201771@qq.com)

package sdp

import (
	"strconv"
	"strings"
)

type RawContext struct {
	MediaDescList []MediaDesc
}

type MediaDesc struct {
	M         M
	ARtpMap   ARtpMap
	AFmtPBase *AFmtPBase
	AControl  AControl
}

type M struct {
	Media string
}

type ARtpMap struct {
	PayloadType        int
	EncodingName       string
	ClockRate          int
	EncodingParameters string
}

type AFmtPBase struct {
	Format     int               // same as PayloadType
	Parameters map[string]string // name -> value
}

type AControl struct {
	Value string
}

// 例子见单元测试
func ParseSdp2RawContext(b []byte) (RawContext, error) {
	var (
		sdpCtx RawContext
		md     *MediaDesc
	)

	s := string(b)
	lines := strings.Split(s, "\r\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "m=") {
			m, err := ParseM(line)
			if err != nil {
				return sdpCtx, err
			}
			if md != nil {
				sdpCtx.MediaDescList = append(sdpCtx.MediaDescList, *md)
			}
			md = &MediaDesc{
				M: m,
			}
		}
		if strings.HasPrefix(line, "a=rtpmap") {
			aRtpMap, err := ParseARtpMap(line)
			if err != nil {
				return sdpCtx, err
			}
			if md == nil {
				continue
			}
			md.ARtpMap = aRtpMap
		}
		if strings.HasPrefix(line, "a=fmtp") {
			aFmtPBase, err := ParseAFmtPBase(line)
			if err != nil {
				return sdpCtx, err
			}
			if md == nil {
				continue
			}
			md.AFmtPBase = &aFmtPBase
		}
		if strings.HasPrefix(line, "a=control") {
			aControl, err := ParseAControl(line)
			if err != nil {
				return sdpCtx, err
			}
			if md == nil {
				continue
			}
			md.AControl = aControl
		}
	}
	if md != nil {
		sdpCtx.MediaDescList = append(sdpCtx.MediaDescList, *md)
	}

	return sdpCtx, nil
}

func ParseM(s string) (ret M, err error) {
	ss := strings.TrimPrefix(s, "m=")
	items := strings.Split(ss, " ")
	if len(items) < 1 {
		return ret, ErrSdp
	}
	ret.Media = items[0]
	return
}

// 例子见单元测试
func ParseARtpMap(s string) (ret ARtpMap, err error) {
	// rfc 3640 3.3.1.  General
	// rfc 3640 3.3.6.  High Bit-rate AAC
	//
	// a=rtpmap:<payload type> <encoding name>/<clock rate>[/<encoding parameters>]
	//

	items := strings.SplitN(s, ":", 2)
	if len(items) != 2 {
		err = ErrSdp
		return
	}
	items = strings.SplitN(items[1], " ", 2)
	if len(items) != 2 {
		err = ErrSdp
		return
	}
	ret.PayloadType, err = strconv.Atoi(items[0])
	if err != nil {
		return
	}
	items = strings.SplitN(items[1], "/", 3)
	switch len(items) {
	case 3:
		ret.EncodingParameters = items[2]
		fallthrough
	case 2:
		ret.EncodingName = items[0]
		ret.ClockRate, err = strconv.Atoi(items[1])
		if err != nil {
			return
		}
	default:
		err = ErrSdp
	}
	return
}

// 例子见单元测试
func ParseAFmtPBase(s string) (ret AFmtPBase, err error) {
	// rfc 3640 4.4.1.  The a=fmtp Keyword
	//
	// a=fmtp:<format> <parameter name>=<value>[; <parameter name>=<value>]
	//

	ret.Parameters = make(map[string]string)

	items := strings.SplitN(s, ":", 2)
	if len(items) != 2 {
		err = ErrSdp
		return
	}

	items = strings.SplitN(items[1], " ", 2)
	if len(items) != 2 {
		err = ErrSdp
		return
	}

	ret.Format, err = strconv.Atoi(items[0])
	if err != nil {
		return
	}

	items = strings.Split(items[1], ";")
	for _, pp := range items {
		pp = strings.TrimSpace(pp)
		if len(pp) <= 0 {
			continue
		}
		kv := strings.SplitN(pp, "=", 2)
		if len(kv) != 2 {
			err = ErrSdp
			return
		}
		ret.Parameters[kv[0]] = kv[1]
	}

	return
}

func ParseAControl(s string) (ret AControl, err error) {
	if !strings.HasPrefix(s, "a=control:") {
		err = ErrSdp
		return
	}
	ret.Value = strings.TrimPrefix(s, "a=control:")
	return
}
