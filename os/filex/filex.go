// Copyright 2018 The Teamlint Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.
// You can obtain one at https://github.com/teamlint/go.

// Copyright 2017 gf Author(https://gitee.com/johng/gf). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://gitee.com/johng/gf.

// Package filex 文件管理.
package filex

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

// Mkdir 给定文件的绝对路径创建文件
func Mkdir(fpath string) error {
	err := os.MkdirAll(fpath, os.ModePerm) // 生成多级目录
	return err
}

// Create 给定文件的绝对路径创建文件
func Create(filename string, src ...io.Reader) error {
	dir := Dir(filename)
	var err error
	if !Exists(dir) {
		err = Mkdir(dir)
		if err != nil {
			return err
		}
	}
	if len(src) > 0 {
		out, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			return err
		}
		defer out.Close()
		_, err = io.Copy(out, src[0])
		return err
	}
	// 单纯创建文件
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return nil
}

// Open 打开文件
func Open(path string, pflag ...int) (*os.File, error) {
	flag := os.O_RDWR | os.O_CREATE
	if len(pflag) > 0 {
		flag = pflag[0]
	}

	f, err := os.OpenFile(path, flag, 0666)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// Exists 判断所给路径文件/文件夹是否存在
func Exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	}
	return true
}

// IsDir 判断所给路径是否为文件夹
func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// IsFile 判断所给路径是否为文件
func IsFile(path string) bool {
	return !IsDir(path)
}

// Info 获取文件或目录信息
func Info(path string) *os.FileInfo {
	info, err := os.Stat(path)
	if err != nil {
		return nil
	}
	return &info
}

// MTime 修改时间(秒)
func MTime(path string) int64 {
	f, e := os.Stat(path)
	if e != nil {
		return 0
	}
	return f.ModTime().Unix()
}

// MTimeMS 修改时间(毫秒)
func MTimeMS(path string) int64 {
	f, e := os.Stat(path)
	if e != nil {
		return 0
	}
	return int64(f.ModTime().Nanosecond() / 1000000)
}

// Size 文件大小(bytes)
func Size(path string) int64 {
	f, e := os.Stat(path)
	if e != nil {
		return 0
	}
	return f.Size()
}

// ReadableSize 可读性强的文件大小字符串
func ReadableSize(path string) string {
	return FormatSize(float64(Size(path)))
}

// FormatSize 格式化文件大小
func FormatSize(raw float64) string {
	var t float64 = 1024
	var d float64 = 1

	if raw < t {
		return fmt.Sprintf("%.2fB", raw/d)
	}

	d *= 1024
	t *= 1024

	if raw < t {
		return fmt.Sprintf("%.2fK", raw/d)
	}

	d *= 1024
	t *= 1024

	if raw < t {
		return fmt.Sprintf("%.2fM", raw/d)
	}

	d *= 1024
	t *= 1024

	if raw < t {
		return fmt.Sprintf("%.2fG", raw/d)
	}

	d *= 1024
	t *= 1024

	if raw < t {
		return fmt.Sprintf("%.2fT", raw/d)
	}

	d *= 1024
	t *= 1024

	if raw < t {
		return fmt.Sprintf("%.2fP", raw/d)
	}

	return "TooLarge"
}

// Move 文件移动/重命名
func Move(src string, dst string) error {
	dir := Dir(dst)
	if !Exists(dir) {
		err := Mkdir(dir)
		if err != nil {
			return err
		}
	}
	return os.Rename(src, dst)
}

// Rename 文件移动/重命名
func Rename(src string, dst string) error {
	return Move(src, dst)
}

// Copy 文件复制
func Copy(src string, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	dir := Dir(dst)
	if !Exists(dir) {
		err := Mkdir(dir)
		if err != nil {
			return err
		}
	}
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}
	err = dstFile.Sync()
	if err != nil {
		return err
	}
	srcFile.Close()
	dstFile.Close()
	return nil
}

// Glob 文件名正则匹配查找
func Glob(pattern string) ([]string, error) {
	return filepath.Glob(pattern)
}

// Remove 文件/目录删除
func Remove(path string) error {
	return os.RemoveAll(path)
}

// IsReadable 文件是否可读
func IsReadable(path string) bool {
	result := true
	file, err := os.OpenFile(path, os.O_RDONLY, 0666)
	if err != nil {
		result = false
	}
	file.Close()
	return result
}

// IsWritable 文件是否可写
func IsWritable(path string) bool {
	result := true
	if IsDir(path) {
		// 如果是目录，那么创建一个临时文件进行写入测试
		tfile := strings.TrimRight(path, string(filepath.Separator)) + string(filepath.Separator) + string(time.Now().UnixNano())
		err := Create(tfile)
		if err != nil || !Exists(tfile) {
			result = false
		} else {
			Remove(tfile)
		}
	} else {
		// 如果是文件，那么判断文件是否可打开
		file, err := os.OpenFile(path, os.O_WRONLY, 0666)
		if err != nil {
			result = false
		}
		file.Close()
	}
	return result
}

// Chmod 修改文件/目录权限
func Chmod(path string, mode os.FileMode) error {
	return os.Chmod(path, mode)
}

// ScanDir 打开目录，并返回其下一级子目录名称列表，按照文件名称大小写进行排序
func ScanDir(path string) []string {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}

	list, err := f.Readdirnames(-1)
	f.Close()
	if err != nil {
		return nil
	}
	sort.Slice(list, func(i, j int) bool { return list[i] < list[j] })
	return list
}

// RealPath 将所给定的路径转换为绝对路径
// 并判断文件路径是否存在，如果文件不存在，那么返回空字符串
func RealPath(path string) string {
	p, err := filepath.Abs(path)
	if err != nil {
		return ""
	}
	if !Exists(p) {
		return ""
	}
	return p
}

// GetContents (文本)读取文件内容
func GetContents(path string) string {
	return string(GetBinContents(path))
}

// GetBinContents (二进制)读取文件内容
func GetBinContents(path string) []byte {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil
	}
	return data
}

// putContents 写入文件内容
func putContents(path string, data []byte, flag int, perm os.FileMode) error {
	// 支持目录递归创建
	dir := Dir(path)
	if !Exists(dir) {
		if err := Mkdir(dir); err != nil {
			return err
		}
	}
	// 创建/打开文件
	f, err := os.OpenFile(path, flag, perm)
	if err != nil {
		return err
	}
	defer f.Close()
	n, err := f.Write(data)
	if err != nil {
		return err
	} else if n < len(data) {
		return io.ErrShortWrite
	}
	return nil
}

// Truncate changes the size of the named file.
func Truncate(path string, size int) error {
	return os.Truncate(path, int64(size))
}

// PutContents (文本)写入文件内容
func PutContents(path string, content string) error {
	return putContents(path, []byte(content), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
}

// AppendContents (文本)追加内容到文件末尾
func AppendContents(path string, content string) error {
	return putContents(path, []byte(content), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
}

// PutBinContents (二进制)写入文件内容
func PutBinContents(path string, content []byte) error {
	return putContents(path, content, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
}

// AppendBinContents (二进制)追加内容到文件末尾
func AppendBinContents(path string, content []byte) error {
	return putContents(path, content, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
}

// ExecPath 获取当前执行文件的绝对路径
func ExecPath() string {
	p, _ := filepath.Abs(os.Args[0])
	return p
}

// ExecDir 获取当前执行文件的目录绝对路径
func ExecDir() string {
	return filepath.Dir(ExecPath())
}

// Filename Basename 别名
// Basename 获取指定文件路径的文件名称
var Filename = Basename

// Basename 获取指定文件路径的文件名称
func Basename(path string) string {
	return filepath.Base(path)
}

// Dir 获取指定文件路径的目录地址绝对路径
func Dir(filename string) string {
	return filepath.Dir(filename)
}

// Ext 获取指定文件路径的文件扩展名
func Ext(path string) string {
	return filepath.Ext(path)
}

// Home 获取用户主目录
func Home() (string, error) {
	u, err := user.Current()
	if nil == err {
		return u.HomeDir, nil
	}
	if "windows" == runtime.GOOS {
		return homeWindows()
	}
	return homeUnix()
}

func homeUnix() (string, error) {
	if home := os.Getenv("HOME"); home != "" {
		return home, nil
	}
	var stdout bytes.Buffer
	cmd := exec.Command("sh", "-c", "eval echo ~$USER")
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return "", err
	}

	result := strings.TrimSpace(stdout.String())
	if result == "" {
		return "", errors.New("blank output when reading home directory")
	}

	return result, nil
}

func homeWindows() (string, error) {
	drive := os.Getenv("HOMEDRIVE")
	path := os.Getenv("HOMEPATH")
	home := drive + path
	if drive == "" || path == "" {
		home = os.Getenv("USERPROFILE")
	}
	if home == "" {
		return "", errors.New("HOMEDRIVE, HOMEPATH, and USERPROFILE are blank")
	}

	return home, nil
}

// GetNextCharOffset 获得文件内容下一个指定字节的位置
func GetNextCharOffset(file *os.File, char string, start int64) int64 {
	c := []byte(char)[0]
	b := make([]byte, 1)
	o := start
	for {
		_, err := file.ReadAt(b, o)
		if err != nil {
			return 0
		}
		if b[0] == c {
			return o
		}
		o++
	}
}

// GetBinContentByTwoOffsets 获得文件内容中两个offset之间的内容 [start, end)
func GetBinContentByTwoOffsets(file *os.File, start int64, end int64) []byte {
	buffer := make([]byte, end-start)
	if _, err := file.ReadAt(buffer, start); err != nil {
		return nil
	}
	return buffer
}

// TempDir 系统临时目录
func TempDir() string {
	return os.TempDir()
}
