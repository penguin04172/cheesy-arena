Switch Modify for Community Event
=========
## 說明

官方與其他系統都是使用現成的網管型 Vlan Switch(注意：要網管跟Vlan)
沒錢，所以一樣用OpenWRT軟路由完成這項工作
此版本程式也將Switch.go程式調整完畢，使用上沒遇到問題
建議有兩個以上網路口，一個接外網，一個接AP，有多孔可以考慮接有線到DS
網路卡比較便宜的可以考慮 瑞昱 RTL 8111系列，我是使用RTL 8111H *9
要比較穩定可以考慮 Intel 82576/I219等系列

作業系統目前只有用Windows測試過，測試過後建議將外網連線的部分獨立，但沒測試過，所以是以整合的形式說明

## 硬體裝置

本次測試用的主機配置如下
| 類型 | 型號|
| --- | --- |
| CPU | AMD Ryzen PRO R3-4350G |
| 主機板 | MSI B550M Pro-VDH |
| 記憶體 | SK-Hynex DDR4 2666 16G *2 |
| 硬碟 | Intel 660p 512G M.2 PCIe SSD |
| 顯示卡 | Intel Irix Xe (80EU) 4GB |
| 網路卡 | RTL 8111H 4 port *2 |
| 電源 | Montech Beta 550W Bronze |


## 前置準備
1. 下載 [Rufus](https://rufus.ie/zh_TW/)
2. 下載 [OpenWRT](https://firmware-selector.openwrt.org/?version=23.05.2&target=x86%2F64&id=generic) (建議使用Combined-efi版本)
3. 打開 **磁碟管理**>**動作**>**建立 VHD**
    3.1. 選擇要儲存的位置與名稱
    3.2. 容量設置為2GB以上
    3.3. 型態為**VHDX**，大小**動態配置**
4. 打開 Rufus
    4.1. 硬碟 是剛剛新增的 **No Label**
    4.2. 選擇 **OpenWRT韌體**
    4.3. 執行
5. 退出硬碟
    5.1. 打開 **磁碟管理**
    5.2. 右鍵 **剛剛新增的VHDX**>**中斷連結**>**確定**

## OpenWRT安裝

**開始前請確認BIOS中已打開 CPU虛擬化(Intel VT-d/x, AMD SVM)與IO虛擬化 (SRIOV, IOMMU)**

### Windows
1. 打開 **控制台>程式集>開啟或關閉Windows功能**，勾選**Hyper-V**，確定並等待安裝
2. 打開 Hyper-V管理員(有一說一，Hyper-V的網路卡超級難設定)
    2.1. 打開 虛擬交換器管理員
        2.1.1. 新增 **外部網路**，網路介面卡選擇 **你的外網接口**，底下 **取消勾選** 允許管理作業系統共用此介面卡
        2.1.2. 重複 2.1.1. 新增 **連接AP的接口**
        2.1.3. 新增 **內部網路**
3. 新增 > 虛擬機器
    3.1. 設定 名稱與儲存位置
    3.2. 選擇 第二代機器
    3.3. 設定 記憶體大小
    3.4. 跳過 網路介面卡設定
    3.5. 選擇 **使用現有的虛擬硬碟**>**前置作業中燒錄好的VHDX**
    3.6. 完成 並不啟動機器
4. 虛擬機器 > 設置
    4.1. 關閉 **安全啟動**
    4.2. 設定 **主機關機時關閉虛擬機器電源**
    4.3. 設定 網路介面卡
        4.3.1. 以系統管理員身分 打開 PowerShell
        4.3.2. PowerShell 輸入 `Get-VMNetworkAdapter -VMName [VMName]` 取得虛擬機網路卡
        4.3.2. 虛擬機設定 將原本的網路介面卡裝置修改為**內部網路裝置**
        4.3.3. 進階設定 > 勾選 **允許修改MAC位置** 並套用設定
        4.3.4. PowerShell 輸入 `Rename-VMNetworkAdapter -VMNetworkAdapter 網路介面卡 -NewName [NewName]` 修改網路介面卡名稱
        4.3.5. 新增網路介面卡，設置為**外網裝置**，進行4.3.3.與4.3.4.
        4.3.6. 新增網路介面卡，設置為**AP裝置**，進行4.3.3.與4.3.4.
        4.3.7. PowerShell 輸入 `Set-VMNetworkAdapterVlan -VMName [VMName] -VMNetworkAdapterName [AP介面卡名稱] -Trunk -AllowedVlanIdList 1-4094 -NativeVlanId 100`
        4.3.8. 確認虛擬機設定，關閉PowerShell
    4.4. 啟動 虛擬機器

## OpenWRT 設定

### 軟體安裝
1. 與 AP 相同，建議安裝中文包與Argon

### 介面設定
1. 介面 > 裝置 > 編輯 **br-lan**
    3.1. 設定 **接口** 為 **所有LAN裝置**
    3.2. 設定 VLAN
    3.2.1. 新增 VLAN 10-60, 100
    3.2.2. 設置 內部網路裝置 為 Vlan 100, Untagged, Primary
    3.2.3. 設置 AP裝置 為 Vlan 10-60, Tagged
    3.2.4. 設置 AP裝置 為 Vlan 100, Tagged, Primary
    3.2.5. 其他裝置依照需求設定為 Vlan Untagged, Primary
2. 介面 > 介面 > 編輯 **lan**
    2.1. 設定 裝置為 **br-lan.100**
    2.2. 設定 IP為 **10.0.100.1**，遮罩 **255.255.255.0**，DHCP **禁用** DHCP6與RA
3. 介面 > 介面
    3.1. 新增 介面 **Vlan 10**
    3.2. 設定 裝置 **br-lan.10**，協議 **靜態**
    3.3. 設定 IP **10.0.1.4**，遮罩 **255.255.255.0**，防火牆 **新增 Vlan10**，安裝 DHCP 伺服器
    3.4. 重複前面步驟 新增 Vlan 20-60
4. 儲存並應用

### 防火牆設定
1. 防火牆 > 流量規則
    1.1. 拒絕 TCP VLAN 10-60 到 LAN 10.0.100.5 port 80,443 IPv4
    1.2. 接受 所有 VLAN 10-60 到 LAN 10.0.100.5 IPv4
    1.3. 接受 所有 LAN 10.0.100.5 到 VLAN 10-60 IPv4
    1.4. 拒絕 TCP VLAN 10-60 到 裝置 port 80,443 IPv4