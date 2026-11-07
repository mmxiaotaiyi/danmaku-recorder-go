package cli

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	bili "dm_client_go"
)

func (a *App) cmdSetCookies(line string) {
	rest := ""
	if i := strings.IndexAny(line, " \t"); i >= 0 {
		rest = strings.TrimSpace(line[i+1:])
	}
	if rest == "" {
		fmt.Println("Cookie不能为空")
		return
	}
	bili.SetHeaderCookie(rest)
	fmt.Println("Cookie已更新")
	uid := extractDedeUserID(rest)
	if uid > 0 {
		bili.SetUID(uid)
		fmt.Printf("已提取并设置UID: %d\n", uid)
		return
	}
	fmt.Println("未能从Cookie中提取到有效的DedeUserID")
}

func extractDedeUserID(cookies string) int64 {
	re := regexp.MustCompile(`DedeUserID=(\d+)`)
	matches := re.FindStringSubmatch(cookies)
	if len(matches) >= 2 {
		uid, err := strconv.ParseInt(matches[1], 10, 64)
		if err == nil {
			return uid
		}
	}
	return 0
}
