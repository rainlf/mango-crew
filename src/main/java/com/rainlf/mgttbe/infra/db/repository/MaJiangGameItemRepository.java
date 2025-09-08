package com.rainlf.mgttbe.infra.db.repository;

import com.rainlf.mgttbe.infra.aop.ExecutionTime;
import com.rainlf.mgttbe.infra.db.dataobj.MaJiangGameItemDO;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.List;

@Repository
public interface MaJiangGameItemRepository extends JpaRepository<MaJiangGameItemDO, Integer> {

    List<MaJiangGameItemDO> findByGameIdAndType(Integer gameId, Integer type);

    List<MaJiangGameItemDO> findByGameIdInAndUserIdAndType(List<Integer> gameIds, Integer userId, Integer type);

    List<MaJiangGameItemDO> findByGameIdInAndUserId(List<Integer> gameIds, Integer userId);

    List<MaJiangGameItemDO> findByGameId(Integer gameIds);

    @Query("select m.gameId from MaJiangGameItemDO m where m.userId = :userId order by m.createdTime desc limit :limit")
    List<Integer> findLastGameIdsByUserId(@Param("userId") Integer userId, @Param("limit") Integer limit);

    @ExecutionTime
    @Query("select m.gameId from MaJiangGameItemDO m where m.userId = :userId and m.type in (:types) order by m.createdTime desc limit :limit offset :offset")
    List<Integer> findLastGameIdsByUserIdAndTypeIn(@Param("userId") Integer userId, @Param("types") List<Integer> types, @Param("limit") Integer limit, @Param("offset") Integer offset);

}
