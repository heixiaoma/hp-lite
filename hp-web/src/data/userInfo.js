export default {
    KEY: "USER_INFO",
    setUserInfo(userInfo) {
        localStorage.setItem(this.KEY, JSON.stringify(userInfo))
    },
    getUserInfo() {
        try {
            return JSON.parse(localStorage.getItem(this.KEY))
        } catch (e) {
            return null
        }
    },
    removeUserInfo(){
        localStorage.removeItem(this.KEY)
    }
}
