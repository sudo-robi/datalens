package datalens;

import datalens.model.AuthDTO;
import datalens.service.AuthService;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.test.context.ActiveProfiles;
import org.springframework.transaction.annotation.Transactional;

import static org.junit.jupiter.api.Assertions.*;

@SpringBootTest
@ActiveProfiles("test")
@Transactional
class AuthServiceTest {

    @Autowired
    private AuthService authService;

    @Test
    void registerCreatesNewUser() {
        AuthDTO.RegisterRequest request = new AuthDTO.RegisterRequest();
        request.setUsername("testuser");
        request.setEmail("test@example.com");
        request.setPassword("password123");

        AuthDTO.AuthResponse response = authService.register(request);

        assertNotNull(response.getToken());
        assertEquals("testuser", response.getUsername());
        assertEquals("test@example.com", response.getEmail());
    }

    @Test
    void registerDuplicateUsernameThrows() {
        AuthDTO.RegisterRequest request = new AuthDTO.RegisterRequest();
        request.setUsername("duplicate");
        request.setEmail("first@example.com");
        request.setPassword("password123");
        authService.register(request);

        AuthDTO.RegisterRequest duplicate = new AuthDTO.RegisterRequest();
        duplicate.setUsername("duplicate");
        duplicate.setEmail("second@example.com");
        duplicate.setPassword("password123");

        assertThrows(RuntimeException.class, () -> authService.register(duplicate));
    }

    @Test
    void loginReturnsToken() {
        AuthDTO.RegisterRequest registerRequest = new AuthDTO.RegisterRequest();
        registerRequest.setUsername("logintest");
        registerRequest.setEmail("login@example.com");
        registerRequest.setPassword("password123");
        authService.register(registerRequest);

        AuthDTO.LoginRequest loginRequest = new AuthDTO.LoginRequest();
        loginRequest.setUsername("logintest");
        loginRequest.setPassword("password123");

        AuthDTO.AuthResponse response = authService.login(loginRequest);

        assertNotNull(response.getToken());
        assertEquals("logintest", response.getUsername());
    }
}
