package net.hserver.hplite.domian.bean;

import lombok.Data;

@Data
public class ResUserKey {
    private String key;

    private String desc;

    private Integer userId;
    private String username;
    private String userDesc;
}
