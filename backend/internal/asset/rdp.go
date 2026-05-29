package asset

import (
	"fmt"
	"net/http"
	"strconv"
)

func handleRDP(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		respondErr(w, http.StatusBadRequest, "missing asset id")
		return
	}
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		respondErr(w, http.StatusBadRequest, "invalid id")
		return
	}

	a, err := repo.GetAsset(r.Context(), id)
	if err != nil || a == nil {
		respondErr(w, http.StatusNotFound, "asset not found")
		return
	}

	port := 3389
	if p, err := strconv.Atoi(a.RDPPort); err == nil && p > 0 {
		port = p
	}

	rdpContent := fmt.Sprintf("full address:s:%s:%d\r\nserver port:i:%d\r\nusername:s:%s\r\nscreen mode id:i:2\r\ndesktopwidth:i:1920\r\ndesktopheight:i:1080\r\nsession bpp:i:32\r\ncompression:i:1\r\nkeyboardhook:i:2\r\naudiomode:i:0\r\ndisableremoteappcapscheck:i:1\r\nautoreconnection enabled:i:1\r\nauthentication level:i:2\r\nprompt for credentials:i:1\r\nnegotiate security layer:i:1\r\nremoteapplicationmode:i:0\r\nalternate shell:s:\r\nshell working directory:s:\r\ngatewayhostname:s:\r\ngatewayusagemethod:i:4\r\ngatewaycredentialssource:i:4\r\ngatewayprofileusagemethod:i:0\r\npromptcredentialonce:i:1\r\nuse redirection server name:i:0\r\n", a.IP, port, port, a.RDPUser)

	w.Header().Set("Content-Type", "application/x-rdp")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.rdp\"", a.Name))
	w.Write([]byte(rdpContent))
}
