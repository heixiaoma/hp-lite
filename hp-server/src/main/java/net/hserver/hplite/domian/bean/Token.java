package net.hserver.hplite.domian.bean;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;
import lombok.ToString;

@Data
@AllArgsConstructor
@NoArgsConstructor
@ToString
public class Token {

    private Integer userId;

    private Role role;

    private Long time;

    public enum Role {
        CLIENT, ADMIN
    }

}
