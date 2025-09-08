package com.rainlf.mgttbe.infra.db.dataobj;

import jakarta.persistence.*;
import lombok.Data;

import java.time.LocalDateTime;

@Data
@Entity
@Table(name = "mgtt_majiang_game")
public class MaJiangGameDO {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Integer id;
    private Integer type;
    private Integer player1;
    private Integer player2;
    private Integer player3;
    private Integer player4;
    private Integer isDeleted;

    @Column(insertable = false, updatable = false, columnDefinition = "datetime default CURRENT_TIMESTAMP not null")
    private LocalDateTime createdTime;
    @Column(insertable = false, updatable = false, columnDefinition = "datetime default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP")
    private LocalDateTime updatedTime;
}
