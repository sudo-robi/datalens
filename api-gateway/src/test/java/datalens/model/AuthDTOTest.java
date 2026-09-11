package datalens.model;

import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.*;

class AuthDTOTest {

    @Test
    void registerRequestBuilderAndGetters() {
        AuthDTO.RegisterRequest request = new AuthDTO.RegisterRequest();
        request.setUsername("testuser");
        request.setEmail("test@example.com");
        request.setPassword("securepass123");

        assertEquals("testuser", request.getUsername());
        assertEquals("test@example.com", request.getEmail());
        assertEquals("securepass123", request.getPassword());
    }

    @Test
    void loginRequestBuilderAndGetters() {
        AuthDTO.LoginRequest request = new AuthDTO.LoginRequest();
        request.setUsername("loginuser");
        request.setPassword("mypassword");

        assertEquals("loginuser", request.getUsername());
        assertEquals("mypassword", request.getPassword());
    }

    @Test
    void authResponseBuilder() {
        AuthDTO.AuthResponse response = AuthDTO.AuthResponse.builder()
                .token("jwt-token-xyz")
                .username("respuser")
                .email("resp@example.com")
                .build();

        assertEquals("jwt-token-xyz", response.getToken());
        assertEquals("respuser", response.getUsername());
        assertEquals("resp@example.com", response.getEmail());
    }

    @Test
    void registerRequestSettersAndGettersRoundTrip() {
        AuthDTO.RegisterRequest req = new AuthDTO.RegisterRequest();
        req.setUsername("abc");
        req.setEmail("a@b.com");
        req.setPassword("pw");

        assertEquals("abc", req.getUsername());
        assertEquals("a@b.com", req.getEmail());
        assertEquals("pw", req.getPassword());
    }
}
