package com.rainlf.mgttbe.infra.db.dataobj;

import jakarta.persistence.*;
import lombok.Data;

import java.time.LocalDateTime;

@Data
@Entity
@Table(name = "mgtt_majiang_game_item")
public class MaJiangGameItemDO {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Integer id;
    private Integer gameId;
    private Integer userId;
    private Integer type;
    private Integer basePoint;
    private String winTypes;
    private Integer multi;
    private Integer points;

    @Column(insertable = false, updatable = false, columnDefinition = "datetime default CURRENT_TIMESTAMP not null")
    private LocalDateTime createdTime;
    @Column(insertable = false, updatable = false, columnDefinition = "datetime default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP")
    private LocalDateTime updatedTime;
}
