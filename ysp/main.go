package main

import (
	"bytes"
	"flag"
	crand "crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// 全局变量
var (
	TeaCKey, _ = hex.DecodeString("59b2f7cf725ef43c34fdd7c123411ed3")
	XorKey     = []byte{0x84, 0x2E, 0xED, 0x08, 0xF0, 0x66, 0xE6, 0xEA, 0x48, 0xB4, 0xCA, 0xA9, 0x91, 0xED, 0x6F, 0xF3}
	HOST_IP    string // 主机IP地址
)

// 常量定义
const (
	PORT       = "9824" // 固定端口号
	Delta      = uint32(0x9E3779B9)
	Rounds     = 16
	LogRounds  = 4
	SaltLen    = 2
	ZeroLen    = 7
)

// 频道信息结构
type ChannelInfo struct {
	Key     string // 频道标识符
	Name    string // 频道名称
	TvgID   string // EPG频道ID
	Logo    string // logo文件名
	Group   string // 分组
	Cnlid   string // cnlid
	Livepid string // livepid
	Defn    string // 清晰度
}

// 完整的频道列表，按央视频道、卫视频道顺序排列
var channels = []ChannelInfo{
	// 央视频道
	{"cctv1", "CCTV1综合", "CCTV1", "CCTV1.png", "央视频道", "2024078201", "600001859", "fhd"},
	{"cctv2", "CCTV2财经", "CCTV2", "CCTV2.png", "央视频道", "2024075401", "600001800", "fhd"},
	{"cctv3", "CCTV3综艺", "CCTV3", "CCTV3.png", "央视频道", "2024068501", "600001801", "fhd"},
	{"cctv4", "CCTV4中文国际", "CCTV4", "CCTV4.png", "央视频道", "2024078301", "600001814", "fhd"},
	{"cctv5", "CCTV5体育", "CCTV5", "CCTV5.png", "央视频道", "2024078401", "600001818", "fhd"},
	{"cctv5p", "CCTV5+体育赛事", "CCTV5+", "CCTV5+.png", "央视频道", "2024078001", "600001817", "fhd"},
	{"cctv6", "CCTV6电影", "CCTV6", "CCTV6.png", "央视频道", "2013693901", "600108442", "fhd"},
	{"cctv7", "CCTV7国防军事", "CCTV7", "CCTV7.png", "央视频道", "2024072001", "600004092", "fhd"},
	{"cctv8", "CCTV8电视剧", "CCTV8", "CCTV8.png", "央视频道", "2024078501", "600001803", "fhd"},
	{"cctv9", "CCTV9纪录", "CCTV9", "CCTV9.png", "央视频道", "2024078601", "600004078", "fhd"},
	{"cctv10", "CCTV10科教", "CCTV10", "CCTV10.png", "央视频道", "2024078701", "600001805", "fhd"},
	{"cctv11", "CCTV11戏曲", "CCTV11", "CCTV11.png", "央视频道", "2027248701", "600001806", "fhd"},
	{"cctv12", "CCTV12社会与法", "CCTV12", "CCTV12.png", "央视频道", "2027248801", "600001807", "fhd"},
	{"cctv13", "CCTV13新闻", "CCTV13", "CCTV13.png", "央视频道", "2024079001", "600001811", "fhd"},
	{"cctv14", "CCTV14少儿", "CCTV14", "CCTV14.png", "央视频道", "2027248901", "600001809", "fhd"},
	{"cctv15", "CCTV15音乐", "CCTV15", "CCTV15.png", "央视频道", "2027249001", "600001815", "fhd"},
	{"cctv16", "CCTV16奥林匹克", "CCTV16", "CCTV16.png", "央视频道", "2027249101", "600098637", "fhd"},
	{"cctv16_4k", "CCTV16奥林匹克(4K)", "CCTV16", "CCTV16.png", "央视频道", "2027249301", "600099502", "hdr"},
	{"cctv17", "CCTV17农业农村", "CCTV17", "CCTV17.png", "央视频道", "2027249401", "600001810", "fhd"},
	{"cctv4k", "CCTV4K超高清", "CCTV4K", "CCTV4K.png", "央视频道", "2027249501", "600002264", "hdr"},
	{"cctv8k", "CCTV8K超高清", "CCTV8K", "CCTV8K.png", "央视频道", "2026774101", "600156816", "hdr"},
	{"cgtn", "CGTN", "CGTN", "CGTN.png", "央视频道", "2024181701", "600014550", "fhd"},
	{"cgtn_french", "CGTN法语", "CGTN法语", "CGTN法语.png", "央视频道", "2024181801", "600084704", "fhd"},
	{"cgtn_russian", "CGTN俄语", "CGTN俄语", "CGTN俄语.png", "央视频道", "2024181901", "600084758", "fhd"},
	{"cgtn_arabic", "CGTN阿拉伯语", "CGTN阿语", "CGTN阿拉伯语.png", "央视频道", "2024182001", "600084782", "fhd"},
	{"cgtn_spanish", "CGTN西班牙语", "CGTN西语", "CGTN西班牙语.png", "央视频道", "2024182101", "600084744", "fhd"},
	{"cgtn_documentary", "CGTN纪录", "CGTN纪录", "CGTN纪录.png", "央视频道", "2024182301", "600084781", "fhd"},

	// 卫视频道
	{"beijing", "北京卫视", "北京卫视", "北京卫视.png", "卫视频道", "2024052703", "600002309", "fhd"},
	{"dongfang", "东方卫视", "东方卫视", "东方卫视.png", "卫视频道", "2024054503", "600002483", "fhd"},
	{"jiangsu", "江苏卫视", "江苏卫视", "江苏卫视.png", "卫视频道", "2024171103", "600002521", "fhd"},
	{"zhejiang", "浙江卫视", "浙江卫视", "浙江卫视.png", "卫视频道", "2024054703", "600002520", "fhd"},
	{"hunan", "湖南卫视", "湖南卫视", "湖南卫视.png", "卫视频道", "2024054803", "600002475", "fhd"},
	{"hubei", "湖北卫视", "湖北卫视", "湖北卫视.png", "卫视频道", "2024171203", "600002508", "fhd"},
	{"guangdong", "广东卫视", "广东卫视", "广东卫视.png", "卫视频道", "2024060903", "600002485", "fhd"},
	{"guangxi", "广西卫视", "广西卫视", "广西卫视.png", "卫视频道", "2024060703", "600002509", "fhd"},
	{"heilongjiang", "黑龙江卫视", "黑龙江卫视", "黑龙江卫视.png", "卫视频道", "2024061003", "600002498", "fhd"},
	{"hainan", "海南卫视", "海南卫视", "海南卫视.png", "卫视频道", "2024055603", "600002506", "fhd"},
	{"chongqing", "重庆卫视", "重庆卫视", "重庆卫视.png", "卫视频道", "2024061103", "600002531", "fhd"},
	{"shenzhen", "深圳卫视", "深圳卫视", "深圳卫视.png", "卫视频道", "2024061303", "600002481", "fhd"},
	{"sichuan", "四川卫视", "四川卫视", "四川卫视.png", "卫视频道", "2024061403", "600002516", "fhd"},
	{"henan", "河南卫视", "河南卫视", "河南卫视.png", "卫视频道", "2024059703", "600002525", "fhd"},
	{"dongnan", "福建东南卫视", "东南卫视", "东南卫视.png", "卫视频道", "2024061503", "600002484", "fhd"},
	{"guizhou", "贵州卫视", "贵州卫视", "贵州卫视.png", "卫视频道", "2024061603", "600002490", "fhd"},
	{"jiangxi", "江西卫视", "江西卫视", "江西卫视.png", "卫视频道", "2024061703", "600002503", "fhd"},
	{"liaoning", "辽宁卫视", "辽宁卫视", "辽宁卫视.png", "卫视频道", "2024171303", "600002505", "fhd"},
	{"anhui", "安徽卫视", "安徽卫视", "安徽卫视.png", "卫视频道", "2024171403", "600002532", "fhd"},
	{"hebei", "河北卫视", "河北卫视", "河北卫视.png", "卫视频道", "2024171503", "600002493", "fhd"},
	{"shandong", "山东卫视", "山东卫视", "山东卫视.png", "卫视频道", "2024171603", "600002513", "fhd"},
	{"tianjin", "天津卫视", "天津卫视", "天津卫视.png", "卫视频道", "2019927003", "600152137", "fhd"},
	{"shanxi1", "陕西卫视", "陕西卫视", "陕西卫视.png", "卫视频道", "2025561103", "600190400", "fhd"},
	{"neimenggu", "内蒙古卫视", "内蒙古卫视", "内蒙古卫视.png", "卫视频道", "2025561203", "600190401", "fhd"},
	{"gansu", "甘肃卫视", "甘肃卫视", "甘肃卫视.png", "卫视频道", "2025561703", "600190408", "fhd"},
	{"ningxia", "宁夏卫视", "宁夏卫视", "宁夏卫视.png", "卫视频道", "2025608503", "600190737", "fhd"},
	{"shanxi", "山西卫视", "山西卫视", "山西卫视.png", "卫视频道", "2025560803", "600190407", "fhd"},
	{"yunnan", "云南卫视", "云南卫视", "云南卫视.png", "卫视频道", "2025561303", "600190402", "fhd"},
	{"jilin", "吉林卫视", "吉林卫视", "吉林卫视.png", "卫视频道", "2025561503", "600190405", "fhd"},
	{"qinghai", "青海卫视", "青海卫视", "青海卫视.png", "卫视频道", "2025559103", "600190406", "fhd"},
	{"xizang", "西藏卫视", "西藏卫视", "西藏卫视.png", "卫视频道", "2025558003", "600190403", "fhd"},
	{"xinjiang", "新疆卫视", "新疆卫视", "新疆卫视.png", "卫视频道", "2019927403", "600152138", "fhd"},
	{"bingtuan", "兵团卫视", "兵团卫视", "兵团卫视.png", "卫视频道", "2025990501", "600193252", "fhd"},
	{"cetv1", "中国教育电视台1频道", "CETV1", "CETV1.png", "卫视频道", "2022823801", "600171827", "fhd"},
	{"guoxue", "国学频道", "国学频道", "国学频道.png", "卫视频道", "2029360403", "600213139", "fhd"},
}

// 通过Key查找频道信息
func findChannelByKey(key string) *ChannelInfo {
	for _, channel := range channels {
		if channel.Key == key {
			return &channel
		}
	}
	return nil
}

// 自定义base64字母表映射
func customEncode(data []byte) string {
	standardAlphabet := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/="
	customAlphabet := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_-="
	
	encoded := base64.StdEncoding.EncodeToString(data)
	var result strings.Builder
	for _, ch := range encoded {
		if ch == '=' && result.Len()%4 == 0 {
			break
		}
		idx := strings.IndexRune(standardAlphabet, ch)
		if idx >= 0 {
			result.WriteByte(customAlphabet[idx])
		}
	}
	return result.String()
}

// 工具函数
func randomHexStr(length int) string {
	const hexChars = "0123456789ABCDEF"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = hexChars[rand.Intn(len(hexChars))]
	}
	return string(result)
}

func xorArray(data []byte) []byte {
	result := make([]byte, len(data))
	for i := 0; i < len(data); i++ {
		result[i] = data[i] ^ XorKey[i&0xF]
	}
	return result
}

func calcSignature(data []byte) uint32 {
	var signature uint32 = 0
	for _, b := range data {
		signature = (0x83*signature + uint32(b)) & 0x7FFFFFFF
	}
	return signature
}

// TEA加密函数
func teaEncryptECB(plaintext, key []byte) []byte {
	if len(plaintext) != 8 || len(key) != 16 {
		return nil
	}

	y := binary.BigEndian.Uint32(plaintext[:4])
	z := binary.BigEndian.Uint32(plaintext[4:])

	k0 := binary.BigEndian.Uint32(key[:4])
	k1 := binary.BigEndian.Uint32(key[4:8])
	k2 := binary.BigEndian.Uint32(key[8:12])
	k3 := binary.BigEndian.Uint32(key[12:])

	var sum uint32 = 0
	for i := 0; i < Rounds; i++ {
		sum += Delta
		y += ((z << 4) + k0) ^ (z + sum) ^ ((z >> 5) + k1)
		z += ((y << 4) + k2) ^ (y + sum) ^ ((y >> 5) + k3)
	}

	result := make([]byte, 8)
	binary.BigEndian.PutUint32(result[:4], y)
	binary.BigEndian.PutUint32(result[4:], z)
	return result
}

func teaDecryptECB(ciphertext, key []byte) []byte {
	if len(ciphertext) != 8 || len(key) != 16 {
		return nil
	}

	y := binary.BigEndian.Uint32(ciphertext[:4])
	z := binary.BigEndian.Uint32(ciphertext[4:])

	k0 := binary.BigEndian.Uint32(key[:4])
	k1 := binary.BigEndian.Uint32(key[4:8])
	k2 := binary.BigEndian.Uint32(key[8:12])
	k3 := binary.BigEndian.Uint32(key[12:])

	// 使用uint64计算然后转换为uint32
	var sum uint32
	delta64 := uint64(Delta)
	shifted := delta64 << uint64(LogRounds)
	sum = uint32(shifted & 0xFFFFFFFF)
	
	for i := 0; i < Rounds; i++ {
		z -= ((y << 4) + k2) ^ (y + sum) ^ ((y >> 5) + k3)
		y -= ((z << 4) + k0) ^ (z + sum) ^ ((z >> 5) + k1)
		sum -= Delta
	}

	result := make([]byte, 8)
	binary.BigEndian.PutUint32(result[:4], y)
	binary.BigEndian.PutUint32(result[4:], z)
	return result
}

// 加密函数
func encrypt(key, plaintext []byte) []byte {
	inLen := len(plaintext)
	
	// 计算填充后的总长度
	nPadSaltBodyZeroLen := inLen + 1 + SaltLen + ZeroLen
	nPadLen := nPadSaltBodyZeroLen % 8
	if nPadLen != 0 {
		nPadLen = 8 - nPadLen
	}
	totalLen := nPadSaltBodyZeroLen + nPadLen
	
	// 准备输出缓冲区
	outBuf := make([]byte, 0, totalLen)
	
	// 准备状态变量
	srcBuf := make([]byte, 8)
	ivPlain := make([]byte, 8)
	ivCrypt := make([]byte, 8)
	
	// 第一个字节：随机数 + padLen
	srcBuf[0] = byte(rand.Intn(256) & 0xF8) | byte(nPadLen & 0x07)
	srcI := 1
	
	// 填充随机数据
	for nPadLen > 0 && srcI < 8 {
		srcBuf[srcI] = byte(rand.Intn(256))
		srcI++
		nPadLen--
	}
	
	// 处理Salt
	i := 1
	for i <= SaltLen {
		if srcI < 8 {
			srcBuf[srcI] = byte(rand.Intn(256))
			srcI++
			i++
		}
		if srcI == 8 {
			// 加密前异或ivCrypt
			for j := 0; j < 8; j++ {
				srcBuf[j] ^= ivCrypt[j]
			}
			
			// TEA加密
			tempOut := teaEncryptECB(srcBuf, key)
			
			// 加密后异或ivPlain
			for j := 0; j < 8; j++ {
				tempOut[j] ^= ivPlain[j]
			}
			
			// 更新iv
			copy(ivPlain, srcBuf)
			copy(ivCrypt, tempOut)
			
			// 添加到输出
			outBuf = append(outBuf, tempOut...)
			srcI = 0
		}
	}
	
	// 处理明文数据
	inBufIndex := 0
	for inLen > 0 {
		if srcI < 8 {
			srcBuf[srcI] = plaintext[inBufIndex]
			srcI++
			inBufIndex++
			inLen--
		}
		if srcI == 8 {
			// 加密前异或ivCrypt
			for j := 0; j < 8; j++ {
				srcBuf[j] ^= ivCrypt[j]
			}
			
			// TEA加密
			tempOut := teaEncryptECB(srcBuf, key)
			
			// 加密后异或ivPlain
			for j := 0; j < 8; j++ {
				tempOut[j] ^= ivPlain[j]
			}
			
			// 更新iv
			copy(ivPlain, srcBuf)
			copy(ivCrypt, tempOut)
			
			// 添加到输出
			outBuf = append(outBuf, tempOut...)
			srcI = 0
		}
	}
	
	// 处理Zero
	i = 1
	for i <= ZeroLen {
		if srcI < 8 {
			srcBuf[srcI] = 0
			srcI++
			i++
		}
		if srcI == 8 {
			// 加密前异或ivCrypt
			for j := 0; j < 8; j++ {
				srcBuf[j] ^= ivCrypt[j]
			}
			
			// TEA加密
			tempOut := teaEncryptECB(srcBuf, key)
			
			// 加密后异或ivPlain
			for j := 0; j < 8; j++ {
				tempOut[j] ^= ivPlain[j]
			}
			
			// 更新iv
			copy(ivPlain, srcBuf)
			copy(ivCrypt, tempOut)
			
			// 添加到输出
			outBuf = append(outBuf, tempOut...)
			srcI = 0
		}
	}
	
	return outBuf
}

// ckey42生成
func ckey42(platform, timestamp int, sdtfrom, vid, guid, appVer string) string {
	// 构建数据
	var buf bytes.Buffer
	
	// Header
	buf.Write([]byte{0x00, 0x00, 0x00, 0x42, 0x00, 0x00, 0x00, 0x04, 0x00, 0x00, 0x04, 0xd2})
	
	// Platform
	binary.Write(&buf, binary.BigEndian, uint32(platform))
	
	// Signature (4 zero bytes)
	buf.Write([]byte{0x00, 0x00, 0x00, 0x00})
	
	// Timestamp
	binary.Write(&buf, binary.BigEndian, uint32(timestamp))
	
	// 写入字符串字段
	writeString(&buf, sdtfrom)
	
	// randFlag (base64编码的18字节随机数)
	randBytes := make([]byte, 18)
	crand.Read(randBytes)
	writeString(&buf, base64.StdEncoding.EncodeToString(randBytes))
	
	writeString(&buf, appVer)
	writeString(&buf, vid)
	writeString(&buf, guid)
	
	// part1 and isDlna
	binary.Write(&buf, binary.BigEndian, uint32(1))
	binary.Write(&buf, binary.BigEndian, uint32(1))
	
	// uid
	writeString(&buf, "2622783A")
	
	// bundleID
	writeString(&buf, "nil")
	
	// uuid4
	writeString(&buf, newUUID())
	
	// bundleID1
	writeString(&buf, "nil")
	
	// ckeyVersion
	writeString(&buf, "v0.1.000")
	
	// packageName
	writeString(&buf, "com.cctv.yangshipin.app.iphone")
	
	// platform_str
	writeString(&buf, strconv.Itoa(platform))
	
	// ex_json_bus
	writeString(&buf, "ex_json_bus")
	
	// ex_json_vs
	writeString(&buf, "ex_json_vs")
	
	// ck_guard_time
	writeString(&buf, randomHexStr(66))
	
	// 添加长度前缀
	data := buf.Bytes()
	length := len(data)
	lengthBytes := []byte{byte(length >> 8), byte(length & 0xFF)}
	
	finalData := append(lengthBytes, data...)
	
	// 加密
	encrypted := encrypt(TeaCKey, finalData)
	
	// 计算签名
	signature := calcSignature(finalData)
	sigBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(sigBytes, signature)
	encrypted = append(encrypted, sigBytes...)
	
	// XOR处理
	xored := xorArray(encrypted)
	
	// 自定义base64编码
	encoded := customEncode(xored)
	
	return "--01" + encoded
}

func newUUID() string {
	uuidBytes := make([]byte, 16)
	crand.Read(uuidBytes)
	
	// 设置版本号（4）
	uuidBytes[6] = (uuidBytes[6] & 0x0F) | 0x40
	// 设置变体号（RFC 4122）
	uuidBytes[8] = (uuidBytes[8] & 0x3F) | 0x80
	
	return fmt.Sprintf("%x-%x-%x-%x-%x", 
		uuidBytes[0:4], uuidBytes[4:6], uuidBytes[6:8], 
		uuidBytes[8:10], uuidBytes[10:16])
}

func writeString(buf *bytes.Buffer, s string) {
	length := len(s)
	binary.Write(buf, binary.BigEndian, uint16(length))
	buf.WriteString(s)
}

// API响应结构
type YSPResponse struct {
	Playurl string       `json:"playurl"`
	Formats []FormatInfo `json:"formats"`
}

type FormatInfo struct {
	Defn     string `json:"defn"`
	Playurl  string `json:"playurl"`
	Priority int    `json:"priority"`
}

// 获取播放URL
func getPlayURL(cnlid, livepid, defn string) (string, *YSPResponse, error) {
	baseURL := "https://liveinfo.ysp.cctv.cn"
	
	params := url.Values{}
	params.Set("atime", "120")
	params.Set("livepid", livepid)
	params.Set("cnlid", cnlid)
	params.Set("appVer", "V8.22.1035.3031")
	params.Set("app_version", "300090")
	params.Set("caplv", "1")
	params.Set("cmd", "2")
	params.Set("defn", defn)
	params.Set("device", "iPhone")
	params.Set("encryptVer", "4.2")
	params.Set("getpreviewinfo", "0")
	params.Set("hevclv", "33")
	params.Set("lang", "zh-Hans_JP")
	params.Set("livequeue", "0")
	params.Set("logintype", "1")
	params.Set("nettype", "1")
	params.Set("newnettype", "1")
	params.Set("newplatform", "4330403")
	params.Set("platform", "4330403")
	params.Set("playbacktime", "0")
	params.Set("sdtfrom", "v3021")
	params.Set("spacode", "23")
	params.Set("spaudio", "1")
	params.Set("spdemuxer", "6")
	params.Set("spdrm", "2")
	params.Set("spdynamicrange", "7")
	params.Set("spflv", "1")
	params.Set("spflvaudio", "1")
	params.Set("sphdrfps", "60")
	params.Set("sphttps", "0")
	params.Set("spvcode", "MSgzMDoyMTYwLDYwOjIxNjB8MzA6MjE2MCw2MDoyMTYwKTsyKDMwOjIxNjAsNjA6MjE2MHwzMDoyMTYwLDYwOjIxNjAp")
	params.Set("spvideo", "4")
	params.Set("stream", "1")
	params.Set("system", "1")
	params.Set("sysver", "ios18.2.1")
	params.Set("uhd_flag", "4")
	
	// 生成ckey
	platform, _ := strconv.Atoi(params.Get("platform"))
	timestamp := int(time.Now().Unix())
	guid := randomHexStr(32)
	ckey := ckey42(platform, timestamp, "dcgh", cnlid, guid, params.Get("appVer"))
	
	params.Set("cKey", ckey)
	
	// 发送请求
	reqURL := baseURL + "?" + params.Encode()
	
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return "", nil, err
	}
	
	req.Header.Set("User-Agent", "qqlive")
	req.Header.Set("Connection", "Keep-Alive")
	req.Header.Set("Accept-Encoding", "gzip")
	
	resp, err := client.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", nil, err
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", nil, err
	}
	
	// 如果defn是auto，返回格式列表
	if defn == "auto" {
		if formats, ok := result["formats"].([]interface{}); ok {
			var formatList []FormatInfo
			for _, f := range formats {
				if fmtMap, ok := f.(map[string]interface{}); ok {
					format := FormatInfo{
						Defn:    getString(fmtMap, "defn"),
						Playurl: getString(fmtMap, "playurl"),
					}
					if pri, ok := fmtMap["priority"].(float64); ok {
						format.Priority = int(pri)
					}
					formatList = append(formatList, format)
				}
			}
			yspResp := &YSPResponse{Formats: formatList}
			return "", yspResp, nil
		}
	}
	
	// 返回播放URL
	if playurl, ok := result["playurl"].(string); ok {
		return playurl, nil, nil
	}
	
	return "", nil, fmt.Errorf("未找到播放地址")
}

func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// 生成M3U播放列表
func generateM3U() string {
	var buf strings.Builder
	// 添加M3U头部信息
	buf.WriteString(`#EXTM3U x-tvg-url="https://fy.188766.xyz/all.xml.gz"`)
	buf.WriteString("\n")
	
	// 直接遍历排序后的频道列表
	for _, channel := range channels {
		logoURL := fmt.Sprintf("https://gcore.jsdelivr.net/gh/taksssss/tv/icon/%s", channel.Logo)
		
		// 生成符合要求的EXTINF行，tvg-id和tvg-name都使用TvgID
		buf.WriteString(fmt.Sprintf(`#EXTINF:-1 tvg-id="%s" tvg-name="%s" tvg-logo="%s" group-title="%s",%s
`, channel.TvgID, channel.TvgID, logoURL, channel.Group, channel.Name))
		
		// 生成播放URL
		buf.WriteString(fmt.Sprintf("http://%s:%s/ysp?id=%s\n", HOST_IP, PORT, channel.Key))
	}
	
	return buf.String()
}

// HTTP处理器
func yspHandler(w http.ResponseWriter, r *http.Request) {
	// 设置无缓存头
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	
	// 获取id参数
	id := r.URL.Query().Get("id")
	
	// 获取其他参数
	cnlid := r.URL.Query().Get("cnlid")
	livepid := r.URL.Query().Get("livepid")
	defn := r.URL.Query().Get("defn")
	
	// 如果提供了id参数，查找对应的频道信息
	if id != "" {
		// 查找频道信息
		if channel := findChannelByKey(id); channel != nil {
			cnlid = channel.Cnlid
			livepid = channel.Livepid
			if defn == "" {
				defn = channel.Defn
			}
		} else {
			http.Error(w, "频道ID不存在", http.StatusNotFound)
			return
		}
	}
	
	if cnlid == "" || livepid == "" {
		http.Error(w, "缺少cnlid或livepid参数", http.StatusBadRequest)
		return
	}
	
	if defn == "" {
		defn = "auto"
	}
	
	// 获取播放URL
	playurl, formats, err := getPlayURL(cnlid, livepid, defn)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// 如果defn为auto，返回格式列表
	if defn == "auto" {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"formats": formats,
		})
		return
	}
	
	// 重定向到播放URL
	http.Redirect(w, r, playurl, http.StatusFound)
}

func m3uHandler(w http.ResponseWriter, r *http.Request) {
	// 设置响应头
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	
	// 生成M3U内容
	m3uContent := generateM3U()
	
	// 写入响应
	w.Write([]byte(m3uContent))
}

func main() {
	// 解析命令行参数
	flag.StringVar(&HOST_IP, "host", "127.0.0.1", "服务器主机IP地址")
	flag.Parse()
	
	// 检查环境变量，环境变量优先级高于命令行参数
	if envHost := os.Getenv("HOST_IP"); envHost != "" {
		HOST_IP = envHost
	}
	
	// 初始化随机数种子
	rand.Seed(time.Now().UnixNano())
	
	// 设置HTTP路由
	http.HandleFunc("/ysp", yspHandler)
	http.HandleFunc("/ysp.m3u", m3uHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			html := fmt.Sprintf(`<html>
<head><title>央视频道直播代理</title></head>
<body>
<h1>央视频道直播代理服务</h1>
<p>可用端点：</p>
<ul>
<li><a href="/ysp.m3u">/ysp.m3u</a> - 获取M3U播放列表</li>
<li>/ysp?cnlid=频道ID&livepid=直播ID&defn=清晰度 - 获取直播流</li>
<li>/ysp?id=频道名称 - 通过频道名称获取直播流</li>
</ul>
<p>服务器地址：%s:%s</p>
<p>配置方式：</p>
<ol>
<li>命令行参数：--host=IP地址</li>
<li>环境变量：HOST_IP=IP地址</li>
</ol>
</body>
</html>`, HOST_IP, PORT)
			w.Write([]byte(html))
		} else {
			http.NotFound(w, r)
		}
	})
	
	// 启动服务器
	addr := fmt.Sprintf(":%s", PORT)
	server := &http.Server{
		Addr:         addr,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	
	fmt.Printf("服务器启动在 %s:%s\n", HOST_IP, PORT)
	fmt.Printf("访问 http://%s:%s/ysp.m3u 获取播放列表\n", HOST_IP, PORT)
	
	if err := server.ListenAndServe(); err != nil {
		fmt.Printf("服务器启动失败: %v\n", err)
	}
}
