package com.rainlf.mgttbe.model;

import lombok.Getter;

import java.util.Objects;

@Getter
public enum MaJiangGameType {
    PING_HU("平胡", 1),
    ZI_MO("自摸", 2),
    YI_PAO_SHUANG_XIANG("一炮双响", 3),
    YI_PAO_SAN_XIANG("一炮三响", 4),
    XIANG_GONG("相公", 5),
    YUN_DONG("运动", 6),
    ;

    private final String name;
    private final Integer code;

    MaJiangGameType(String name, Integer code) {
        this.name = name;
        this.code = code;
    }

    public static MaJiangGameType fromCode(Integer code) {
        for (MaJiangGameType type : values()) {
            if (Objects.equals(type.code, code)) {
                return type;
            }
        }
        throw new RuntimeException("MaJiangGameType, no such code: " + code);
    }
}
