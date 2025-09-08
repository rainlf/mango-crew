package com.rainlf.mgttbe.infra.db.manager;

import com.fasterxml.jackson.core.type.TypeReference;
import com.rainlf.mgttbe.infra.aop.ExecutionTime;
import com.rainlf.mgttbe.infra.db.dataobj.MaJiangGameItemDO;
import com.rainlf.mgttbe.infra.db.repository.MaJiangGameItemRepository;
import com.rainlf.mgttbe.infra.util.JsonUtils;
import com.rainlf.mgttbe.model.MaJiangGameItem;
import com.rainlf.mgttbe.model.MaJiangUserType;
import com.rainlf.mgttbe.model.MaJiangWinType;
import org.springframework.beans.BeanUtils;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.stream.Collectors;


@Service
public class MaJiangGameItemManager {
    @Autowired
    private MaJiangGameItemRepository majiangGameItemRepository;

    public void save(List<MaJiangGameItem> majiangGameItems) {
        majiangGameItemRepository.saveAll(
                majiangGameItems
                        .stream()
                        .map(this::toMaJiangGameItemDO)
                        .toList()
        );
    }

    public List<MaJiangGameItem> findByGameIdAndUserIdAndType(List<Integer> gameIds, Integer userId, MaJiangUserType userType) {
        return majiangGameItemRepository.findByGameIdInAndUserIdAndType(gameIds, userId, userType.getCode())
                .stream()
                .map(this::toMaJiangGameItem)
                .collect(Collectors.toList());
    }

    public List<MaJiangGameItem> findByGameIdAndType(Integer gameId, MaJiangUserType userType) {
        return majiangGameItemRepository.findByGameIdAndType(gameId, userType.getCode())
                .stream()
                .map(this::toMaJiangGameItem)
                .collect(Collectors.toList());
    }

    public List<MaJiangGameItem> findByGameIdAndUserId(List<Integer> gameIds, Integer userId) {
        return majiangGameItemRepository.findByGameIdInAndUserId(gameIds, userId)
                .stream()
                .map(this::toMaJiangGameItem)
                .collect(Collectors.toList());
    }

    public List<MaJiangGameItem> findByGameId(Integer gameId) {
        return majiangGameItemRepository.findByGameId(gameId)
                .stream()
                .map(this::toMaJiangGameItem)
                .collect(Collectors.toList());
    }

    public List<Integer> findLastGameIdsByUserId(Integer userId, Integer limit) {
        return majiangGameItemRepository.findLastGameIdsByUserId(userId, limit);
    }

    @ExecutionTime
    public List<Integer> findLastGameIdsByUserIdAndTypeIn(Integer userId, List<MaJiangUserType> types,  Integer limit, Integer offset) {
        return majiangGameItemRepository.findLastGameIdsByUserIdAndTypeIn(userId, types.stream().map(MaJiangUserType::getCode).toList(), limit,  offset);
    }

    private MaJiangGameItem toMaJiangGameItem(MaJiangGameItemDO majiangGameItemDO) {
        if (majiangGameItemDO == null) {
            return null;
        }

        MaJiangGameItem majiangGameItem = new MaJiangGameItem();
        BeanUtils.copyProperties(majiangGameItemDO, majiangGameItem);
        majiangGameItem.setType(MaJiangUserType.fromCode(majiangGameItemDO.getType()));
        if (majiangGameItemDO.getWinTypes() != null) {
            majiangGameItem.setWinTypes(JsonUtils.readString(majiangGameItemDO.getWinTypes(), new TypeReference<List<Integer>>() {
            }).stream().map(MaJiangWinType::fromCode).collect(Collectors.toList()));
        }
        return majiangGameItem;
    }

    private MaJiangGameItemDO toMaJiangGameItemDO(MaJiangGameItem maJiangGameItem) {
        if (maJiangGameItem == null) {
            return null;
        }

        MaJiangGameItemDO maJiangGameItemDO = new MaJiangGameItemDO();
        BeanUtils.copyProperties(maJiangGameItem, maJiangGameItemDO);
        maJiangGameItemDO.setType(maJiangGameItem.getType().getCode());
        if (maJiangGameItem.getWinTypes() != null) {
            maJiangGameItemDO.setWinTypes(JsonUtils.writeString(maJiangGameItem.getWinTypes().stream().map(MaJiangWinType::getCode).collect(Collectors.toList())));
        }
        return maJiangGameItemDO;
    }
}
