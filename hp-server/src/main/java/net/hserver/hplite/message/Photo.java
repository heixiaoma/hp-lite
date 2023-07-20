package net.hserver.hplite.message;

import java.util.Arrays;


public class Photo {

    private String username;
    private String domain;

    private PhotoType photoType;

    private byte[] data;

    public Photo(String username,String domain,PhotoType photoType, byte[] data) {
        this.photoType = photoType;
        this.data = data;
        this.domain=domain;
        this.username=username;
    }

    public Photo() {
    }

    public PhotoType getPhotoType() {
        return photoType;
    }

    public void setPhotoType(PhotoType photoType) {
        this.photoType = photoType;
    }

    public byte[] getData() {
        return data;
    }

    public void setData(byte[] data) {
        this.data = data;
    }

    public String getUsername() {
        return username;
    }

    public void setUsername(String username) {
        this.username = username;
    }

    public String getDomain() {
        return domain;
    }

    public void setDomain(String domain) {
        this.domain = domain;
    }

    public enum PhotoType {
        PNG(".png"),
        JPG(".jpg"),
        GIF(".gif");
        private final String tips;

        PhotoType(String tips) {
            this.tips = tips;
        }

        public String getTips() {
            return tips;
        }
    }

    @Override
    public String toString() {
        return "Photo{" +
                "photoType=" + photoType +
                ", data=" + Arrays.toString(data) +
                '}';
    }
}
