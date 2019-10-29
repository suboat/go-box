package box

import (
	"github.com/suboat/go-contrib"
	"github.com/suboat/go-contrib/log"

	"github.com/tudyzhb/yaml"

	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	// 版本号识别
	regVersion   = regexp.MustCompile(`(\w+)?-?v(\d+)\.(\d+)\.(\d+)\(?(\w+)?\)?-?(\w{8})?`)
	regGitCommit = regexp.MustCompile(`v?((\d+)\.(\d+)\.(\d+))?(-\d+)?(-?g?([0-9a-f]+))?(-dirty)?`) // v0.0.1-1-gd4f800c-dirty
)

// 解析版本号
func (v *Version) ParseStr(s string) (err error) {
	if v == nil {
		err = contrib.ErrUndefined
		return
	}
	var (
		vals = regVersion.FindStringSubmatch(s)
	)
	v.Model = vals[1]
	if _v, _err := strconv.ParseUint(vals[2], 10, 32); _err != nil {
		err = contrib.ErrParamInvalid
		return
	} else {
		v.Major = int(_v)
	}
	if _v, _err := strconv.ParseUint(vals[3], 10, 32); _err != nil {
		err = contrib.ErrParamInvalid
		return
	} else {
		v.Minor = int(_v)
	}
	if _v, _err := strconv.ParseUint(vals[4], 10, 32); _err != nil {
		err = contrib.ErrParamInvalid
		return
	} else {
		v.Patch = int(_v)
	}
	if len(vals[5]) > 0 {
		// 优先解析时间
		if _v, _err := time.Parse("01021504", vals[5]); _err == nil {
			v.Build = &_v
		} else {
			_s := vals[5]
			v.Commit = &_s
		}
	}
	if len(vals[6]) > 0 {
		v.Hash = vals[6]
	}
	return
}

// ParseCommit 从gitCommit解析 v0.0.1-1-gd4f800c-dirty
func (v *Version) ParseCommit(gitCommit string) (err error) {
	if len(gitCommit) == 0 {
		return
	}
	var (
		vals   = regGitCommit.FindStringSubmatch(gitCommit)
		ver    = vals[1]
		commit = vals[7]
	)
	if len(ver) == 0 && len(commit) == 0 {
		return
	}

	//
	if len(ver) > 0 {
		//
		if _v, _err := strconv.ParseUint(vals[2], 10, 32); _err == nil {
			v.Major = int(_v)
		} else {
			err = _err
			return
		}
		//
		if _v, _err := strconv.ParseUint(vals[3], 10, 32); _err == nil {
			v.Minor = int(_v)
		} else {
			err = _err
			return
		}
		//
		if _v, _err := strconv.ParseUint(vals[4], 10, 32); _err == nil {
			v.Patch = int(_v)
		} else {
			err = _err
			return
		}
	}

	if len(commit) > 0 {
		if vals[8] == "-dirty" {
			commit += "-dirty"
		}
		v.Commit = &commit
	} else {
		v.Commit = nil
	}
	return
}

//
func (v *Version) ParseInt(i_ int32) (err error) {
	if v == nil {
		err = contrib.ErrUndefined
		return
	}
	i := int(i_)
	v.Major = i / 1000000
	i = i - (v.Major * 1000000)
	v.Minor = i / 1000
	i -= v.Minor * 1000
	v.Patch = i
	return
}

//
func (v *Version) String() (ret string) {
	if v == nil {
		return "v0.0.0"
	}
	if len(v.Model) > 0 {
		ret = v.Model + "-"
	}
	if len(v.Hash) == 0 {
		// 尝试填补Hash
		v.Hash = PubGetRunFileHash()
	}
	if v.Commit == nil && v.Build == nil {
		// 尝试使用修改时间来表示编译时间
		v.Build = PubGetRunFileTime()
	}
	//
	ret = fmt.Sprintf(`%sv%d.%d.%d`, ret, v.Major, v.Minor, v.Patch)
	// 优先输出编译版本
	if v.Commit != nil {
		ret = fmt.Sprintf(`%s(%s)`, ret, *v.Commit)
	} else if v.Build != nil {
		ret = fmt.Sprintf(`%s(%s)`, ret, v.Build.Format("01021504"))
	}
	if len(v.Hash) > 0 {
		ret = fmt.Sprintf(`%s-%s`, ret, v.Hash)
	}
	return
}

//
func (v *Version) Int32() (ret int32) {
	if v == nil {
		return
	}
	ret = int32(v.Major)*1000000 + int32(v.Minor)*1000 + int32(v.Patch)
	return
}

// toJson
func (c *Config) ToJson() (ret string, err error) {
	if c.savePoint == nil {
		err = contrib.ErrConfigSavePointUndefined
		return
	}
	var b []byte
	if b, err = json.Marshal(c.savePoint); err != nil {
		return
	} else {
		ret = string(b)
	}
	return
}

// fromJson
func (c *Config) FromJson(content string) (ret string, err error) {
	c.Lock()
	defer c.Unlock()
	if c.savePoint == nil {
		err = contrib.ErrConfigSavePointUndefined
		return
	}
	if err = json.Unmarshal([]byte(content), c.savePoint); err != nil {
		return
	}
	if ret, err = c.ToJson(); err != nil {
		return
	}
	return
}

// 不打印日志
func (c *Config) SetSilent(v bool) {
	c.silent = v
}

// 保存配置
func (c *Config) Save(newConfig interface{}) (err error) {
	c.Lock()
	defer c.Unlock()
	var (
		savePath    string
		readContent []byte
		saveContent []byte
	)
	if savePath, err = c.GetSavePath(); err != nil {
		return
	}
	// 读旧记录
	readContent, _ = ioutil.ReadFile(savePath)

	if newConfig == nil {
		newConfig = c.savePoint
	}
	if saveContent, err = yaml.MarshalWithComments(newConfig, c.comments); err != nil {
		return
	} else if bytes.Equal(readContent, saveContent) == true {
		// 不重复保存
		return
	}

	// 写入记录
	if err = ioutil.WriteFile(savePath, saveContent, 0666); err != nil {
		return
	}
	if c.silent == false {
		log.Debug(fmt.Sprintf("[config] save config to %s bytes:%d->%d",
			savePath, len(readContent), len(saveContent)))
	}

	// hook
	if c.hookChange != nil {
		if _err := c.hookChange(newConfig); _err != nil {
			log.Warn(fmt.Sprintf(`[config] hookChange error: %v`, _err))
		}
	}
	return
}

// 取保存地址
func (c *Config) GetSavePath() (ret string, err error) {
	if c.pathYaml == nil {
		err = contrib.ErrConfigSavePathUndefined
		return
	} else {
		ret = *c.pathYaml
	}
	return
}

// 设置保存地址
func (c *Config) SetSavePath(savePath string) (err error) {
	if len(savePath) == 0 {
		c.pathYaml = nil
		err = contrib.ErrConfigSavePathUndefined
		return
	} else {
		c.pathYaml = &savePath
	}
	return
}

// 设置要保存的对象
func (c *Config) SetSavePoint(saveTarget interface{}) (err error) {
	c.savePoint = saveTarget
	return
}

// 设置要保存的对象
func (c *Config) SetHookChange(f func(interface{}) error) (err error) {
	c.hookChange = f
	return
}

// 设置备注信息
func (c *Config) SetComments(comments map[string]string) (err error) {
	comments_ := make(map[string]string)
	for key := range comments {
		comments_[strings.ToLower(key)] = comments[key]
	}
	c.comments = comments_
	return
}
