package box

import (
	"github.com/tudyzhb/yaml"

	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

// 读取配置文件
func PubConfigRead(configPath string, targetConfig interface{}, defaultConfig interface{}) (err error) {
	var b []byte
	if b, err = ioutil.ReadFile(configPath); err != nil {
		if os.IsNotExist(err) == false {
			return
		}
		// 自动创建配置文件
		if defaultConfig == nil {
			err = fmt.Errorf(`[config] defaultConfig undefined, "%s" error: %v`, configPath, err)
			return
		}
		if d, _err2 := yaml.Marshal(defaultConfig); _err2 != nil {
			err = _err2
			return
		} else if err = ioutil.WriteFile(configPath, d, 0666); err != nil {
			return
		}
		err = fmt.Errorf(`[config] please modify "%s" and run again`, configPath)
		return
	}

	// 读取配置
	if err = yaml.Unmarshal(b, targetConfig); err != nil {
		return
	}

	// 设置保存地址对对象指针
	type configInterface interface {
		SetSavePath(savePath string) (err error)
		SetSavePoint(saveTarget interface{}) (err error)
	}
	if _c, _ok := targetConfig.(configInterface); _ok == true {
		_c.SetSavePath(configPath)
		_c.SetSavePoint(targetConfig)
	}

	return
}

// 取运行文件的哈希前32位
func PubGetRunFileHash() (ret string) {
	if len(os.Args) > 0 {
		if p, err := filepath.Abs(os.Args[0]); err == nil {
			if f, err := os.Open(p); err == nil {
				defer f.Close()
				h := sha1.New()
				if _, err = io.Copy(h, f); err == nil {
					ret = fmt.Sprintf("%x", h.Sum(nil))
				}
			}
		}
	}
	if len(ret) > 8 {
		ret = ret[0:8]
	}
	return
}

// 取运行文件的修改时间
func PubGetRunFileTime() (ret *time.Time) {
	if len(os.Args) > 0 {
		if p, err := filepath.Abs(os.Args[0]); err == nil {
			if info, err := os.Stat(p); err == nil {
				if t := info.ModTime(); t.Unix() > 0 {
					ret = &t
				}
			}
		}
	}
	return
}
