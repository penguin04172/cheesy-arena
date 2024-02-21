AccessPoint Modify for Community Event
============
### 系統安裝

1. OpenWRT [Table Of Hardware](https://openwrt.org/toh/start)檢查路由器是否支援Openwrt
2. OpenWRT [Firmware Selector](https://firmware-selector.openwrt.org/)下載韌體(Factory)
3. 按照硬體描述頁面更新韌體

### 軟體設定

1. 打開瀏覽器，進入 **管理頁面** 帳號: root 密碼: (空)
2. 如果確定功能完整建議先聯網更新軟體與安裝中文包
2.1. 進入 **Interface(介面)** 頁面確認已連上網路
2.2. 進入 **Software(軟體)** 頁面按下 **更新opkg**
2.3. 檢查 **更新分頁** 將需要更新的軟體進行更新
2.4. 回到 **安裝頁面** 在上方過濾器輸入關鍵字篩選軟體
2.5. 推薦安裝 **luci-i18n-base-zh-tw** 與 **[luci-theme-argon](https://github.com/jerrykuku/luci-theme-argon)**

### 網路設定

注意：請一次設定完畢再按下 **Save & Apply(儲存與應用)**
1. 進入 **Interface(介面)** 頁面
    1.1. 刪除 **Wan**與**Wan6** 介面
    1.2. 進入 **Lan**介面的 **DHCP**分頁
    1.3. 修改 **DHCP6** 與 **RA** 設定為 **Disabled**
2. 進入 **Device(裝置)** 分頁
    2.1. 編輯 **br-lan**
    2.2. 修改 **Bridge Ports(接口)** 選項，選擇 **所有Lan**與**Wan**
    2.3. 進入 **Bridge VLAN filtering(VLAN 過濾)** 分頁
    2.3.1. 新增 VLAN **1, 10, 20, 30, 40, 50, 60, 100**
    2.3.2. 將 **所有LAN** 設定為 VLAN 1 **Untagged**與**Primary(主要)**
    2.3.3. 將 **WAN** 設定為 VLAN 10-60 **Tagged**
    2.3.4. 將 **WAN** 設定為 VLAN 100 **Tagged** 與 **Primary(主要)**
3. 進入 **Interface(介面)** 頁面
    3.1. 編輯 **Lan** 介面，將裝置修改為**br-lan.1**
    3.2. 新增 **VLAN 10**介面，裝置為**br-lan.10**，協定為**Unmanaged(未託管)**
    3.3. 重複步驟 **3.2.** 新增 VLAN 10-60 介面
    3.4. 新增 **VLAN 100**介面，裝置為**br-lan.100**，協定為**Static(靜態)**
    3.5. IP為**10.0.100.3**，遮罩為**255.255.255.0**，上游為**10.0.100.1**
4. **Save & Apply(儲存並應用)**

### 無線設定

注意：可能要配合程式進行修改
推薦：使用第二台無線路由器作為VLAN 100設備，降低負荷
1. 確認是否有5Ghz的radio
2. 設定 2.4Ghz Radio
    2.1. 編輯 2.4Ghz OpenWRT Wifi
    2.1.1. 修改 介面為**Vlan100**
    2.1.2. 修改 名稱與密碼，密碼格式為PSK2
    2.2. 修改 **無線設定**
    2.2.1. 協定為**N**，頻道為**11**，強度為**20**
    2.2.2. 國別碼 **TW**
3. 設定 5Ghz Radio
    2.1. 刪除 5Ghz OpenWRT Wifi
    2.2. 新增 5Ghz Wifi
    2.2.1. 裝置為 **Vlan 10**，名稱為 **1**
    2.2.2. 密碼格式為 **PSK2**，密碼為 **1111111**(隨意，8碼即可)
    2.2.3. 重複 **2.2.1-2** 新增 Wifi 1-6
    2.3. 編輯 **無線設定**
    2.3.1. 協定為 **ac**，頻道為 **157**，強度為 **23**(或更大)
    2.3.2. 國別碼 **TW**
4. **Save & Apply(儲存並應用)**

### 關閉防火牆

推薦：也可以不關慢慢設定，但關閉不影響
1. 進入 **Startup(開機啟動)** 頁面
2. 找到 **Firewall** 項目，右邊 Enable 改為 Disable，按一下 Stop

### 已知問題

1. PlayOff階段，有可能會因為AP導致Arena崩潰，目前確定每場結束都會崩潰，需要修改Access_Point.go程式

### 附錄

20240217 南科模擬賽使用 ASUS AX1800HP(AX54)路由器，依然有負載過大可能性
FIRST 官方使用 LinkSys AC1900WRT

### 注意事項

**明年2025換路由器就不會這麼麻煩了**
目前預估2025換路由器連線架構不會改變，改為FMS與機器人都使用[VH109路由器](https://frc-radio.vivid-hosting.net/)