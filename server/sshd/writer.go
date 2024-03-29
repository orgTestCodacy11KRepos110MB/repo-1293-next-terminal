package sshd

import (
	"encoding/hex"
	"strings"

	"next-terminal/server/api"
	"next-terminal/server/common/term"

	"github.com/gliderlabs/ssh"
)

type Writer struct {
	sessionId    string
	frontSshSess ssh.Session
	backTerm     *term.NextTerminal
	rz           bool
	sz           bool
}

func NewWriter(sessionId string, frontSshSess ssh.Session, backSshSess *term.NextTerminal) *Writer {
	return &Writer{
		sessionId:    sessionId,
		frontSshSess: frontSshSess,
		backTerm:     backSshSess,
	}
}

func (w *Writer) Write(p []byte) (n int, err error) {
	s := string(p)
	if !w.sz && !w.rz {
		// rz的开头字符
		hexData := hex.EncodeToString(p)
		if strings.Contains(hexData, "727a0d2a2a184230303030303030303030303030300d8a11") {
			w.sz = true
		} else if strings.Contains(hexData, "727a2077616974696e6720746f20726563656976652e2a2a184230313030303030303233626535300d8a11") {
			w.rz = true
		}
	}

	if w.sz {
		// sz 会以 OO 结尾
		if "OO" == s {
			w.sz = false
		}
	} else if w.rz {
		// rz 最后会显示 Received /home/xxx
		if strings.Contains(s, "Received") {
			w.rz = false
			// 把上传的文件名称也显示一下
			if w.backTerm.Recorder != nil {
				err := w.backTerm.Recorder.WriteData(s)
				if err != nil {
					return 0, err
				}
			}
			api.SendObData(w.sessionId, s)
		}
	} else {
		if w.backTerm.Recorder != nil {
			err := w.backTerm.Recorder.WriteData(s)
			if err != nil {
				return 0, err
			}
		}
		api.SendObData(w.sessionId, s)
	}

	return w.frontSshSess.Write(p)
}

func (w *Writer) Read(p []byte) (n int, err error) {
	n, err = w.frontSshSess.Read(p)
	return n, err
}
