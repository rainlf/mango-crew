package com.rainlf.mgttbe.service;

import com.rainlf.mgttbe.controller.dto.MaJiangGameLogDTO;
import com.rainlf.mgttbe.controller.dto.PlayersDTO;
import com.rainlf.mgttbe.controller.dto.SaveMaJiangGameRequest;

import java.util.List;

public interface MaJiangService {
    Integer saveMaJiangGame(SaveMaJiangGameRequest request);

    List<MaJiangGameLogDTO> getMaJiangGameLogs(Integer limit, Integer offset);

    List<MaJiangGameLogDTO> getMaJiangGamesByUser(Integer userId, Integer limit, Integer offset);

    void deleteMaJiangGame(Integer id, Integer userId);

    PlayersDTO getMaJiangGamePlayers();
}
