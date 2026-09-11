package datalens.controller;

import com.fasterxml.jackson.databind.ObjectMapper;
import datalens.config.SecurityConfig;
import datalens.model.AuthDTO;
import datalens.service.AuthService;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.web.servlet.WebMvcTest;
import org.springframework.boot.test.mock.mockito.MockBean;
import org.springframework.context.annotation.Import;
import org.springframework.http.MediaType;
import org.springframework.test.web.servlet.MockMvc;

import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.when;
import static org.springframework.security.test.web.servlet.request.SecurityMockMvcRequestPostProcessors.csrf;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.post;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.*;

@WebMvcTest(AuthController.class)
@Import(SecurityConfig.class)
class AuthControllerTest {

    @Autowired
    private MockMvc mockMvc;

    @Autowired
    private ObjectMapper objectMapper;

    @MockBean
    private AuthService authService;

    @MockBean
    private datalens.security.JwtTokenProvider jwtTokenProvider;

    @MockBean
    private datalens.security.CustomUserDetailsService customUserDetailsService;

    @Test
    void registerReturnsToken() throws Exception {
        AuthDTO.AuthResponse response = AuthDTO.AuthResponse.builder()
                .token("jwt-token-abc")
                .username("newuser")
                .email("new@example.com")
                .build();

        when(authService.register(any(AuthDTO.RegisterRequest.class))).thenReturn(response);

        AuthDTO.RegisterRequest request = new AuthDTO.RegisterRequest();
        request.setUsername("newuser");
        request.setEmail("new@example.com");
        request.setPassword("password123");

        mockMvc.perform(post("/api/auth/register")
                        .with(csrf())
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(objectMapper.writeValueAsString(request)))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.token").value("jwt-token-abc"))
                .andExpect(jsonPath("$.username").value("newuser"))
                .andExpect(jsonPath("$.email").value("new@example.com"));
    }

    @Test
    void loginReturnsToken() throws Exception {
        AuthDTO.AuthResponse response = AuthDTO.AuthResponse.builder()
                .token("jwt-token-login")
                .username("existinguser")
                .email("exist@example.com")
                .build();

        when(authService.login(any(AuthDTO.LoginRequest.class))).thenReturn(response);

        AuthDTO.LoginRequest request = new AuthDTO.LoginRequest();
        request.setUsername("existinguser");
        request.setPassword("password123");

        mockMvc.perform(post("/api/auth/login")
                        .with(csrf())
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(objectMapper.writeValueAsString(request)))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.token").value("jwt-token-login"))
                .andExpect(jsonPath("$.username").value("existinguser"));
    }

    @Test
    void registerWithInvalidBodyReturns400() throws Exception {
        AuthDTO.RegisterRequest request = new AuthDTO.RegisterRequest();

        mockMvc.perform(post("/api/auth/register")
                        .with(csrf())
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(objectMapper.writeValueAsString(request)))
                .andExpect(status().isBadRequest());
    }

    @Test
    void loginWithInvalidBodyReturns400() throws Exception {
        AuthDTO.LoginRequest request = new AuthDTO.LoginRequest();

        mockMvc.perform(post("/api/auth/login")
                        .with(csrf())
                        .contentType(MediaType.APPLICATION_JSON)
                        .content(objectMapper.writeValueAsString(request)))
                .andExpect(status().isBadRequest());
    }
}
