package com.rainlf.mgttbe.infra.db.dataobj;

import jakarta.persistence.*;
import lombok.Data;

import java.time.LocalDateTime;

@Data
@Entity
@Table(name = "mgtt_user")
public class UserDO {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Integer id;

    private String username;
    private Integer points;
    private String realName;
    @Lob
    private byte[] avatar;
    private String openId;
    private String sessionKey;
    private Integer isDeleted;
    private LocalDateTime lastLoginTime;

    @Column(insertable = false, updatable = false, columnDefinition = "datetime default CURRENT_TIMESTAMP not null")
    private LocalDateTime createdTime;
    @Column(insertable = false, updatable = false, columnDefinition = "datetime default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP")
    private LocalDateTime updatedTime;
}
