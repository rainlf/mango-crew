package com.rainlf.mgttbe.controller.dto;

import com.rainlf.mgttbe.model.User;
import lombok.Data;
import org.springframework.beans.BeanUtils;
import java.time.LocalDateTime;
import java.util.Base64;
import java.util.List;

@Data
public class UserDTO {
    private Integer id;
    private String username;
    private String avatar;
    private Integer points;
    private List<String> lastTags;
    private LocalDateTime createdTime;
    private LocalDateTime updatedTime;


    public static UserDTO fromUser(User user) {
        UserDTO userDTO = new UserDTO();
        BeanUtils.copyProperties(user, userDTO);
        if (user.getAvatar() != null && user.getAvatar().length > 0) {
            userDTO.setAvatar("data:image/png;base64," + Base64.getEncoder().encodeToString(user.getAvatar()));
        }
        return userDTO;
    }
}
