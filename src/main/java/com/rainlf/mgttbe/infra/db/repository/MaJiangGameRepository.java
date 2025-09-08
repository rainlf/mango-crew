package com.rainlf.mgttbe.infra.db.repository;

import com.rainlf.mgttbe.infra.aop.ExecutionTime;
import com.rainlf.mgttbe.infra.db.dataobj.MaJiangGameDO;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.List;

@Repository
public interface MaJiangGameRepository extends JpaRepository<MaJiangGameDO, Integer> {

    @ExecutionTime
    List<MaJiangGameDO> findByIdIn(List<Integer> ids);

    @Query(value = "select * from mgtt_majiang_game m where m.id = :id and m.is_deleted = 0 for update", nativeQuery = true)
    MaJiangGameDO findByIdWithLock(@Param("id") Integer id);

    @ExecutionTime
    @Query("select m from MaJiangGameDO m where m.isDeleted = 0 order by m.createdTime desc limit :limit")
    List<MaJiangGameDO> findLastGames(@Param("limit") Integer limit);

    @Query("select m from MaJiangGameDO m where (m.player1 = :userId or  m.player2 = :userId or  m.player3 = :userId or  m.player4 = :userId) and m.isDeleted = 0 order by m.createdTime desc limit :limit")
    List<MaJiangGameDO> findLastGamesByUser(@Param("userId") Integer userId, @Param("limit") Integer limit);


    @Query(value = "select * from mgtt_majiang_game m where (m.player1 = :userId or m.player2 = :userId or m.player3 = :userId or m.player4 = :userId) and m.is_deleted = 0 and m.created_time > DATE_SUB(CURRENT_DATE(), INTERVAL :days DAY) order by m.created_time desc", nativeQuery = true)
    List<MaJiangGameDO> findLastGamesByUserAndDays(@Param("userId") Integer userId, @Param("days") Integer days);

}
