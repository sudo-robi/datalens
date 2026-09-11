package datalens.security;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.security.authentication.UsernamePasswordAuthenticationToken;
import org.springframework.security.core.Authentication;
import org.springframework.security.core.userdetails.User;
import org.springframework.security.core.userdetails.UserDetails;

import java.util.Collections;

import static org.junit.jupiter.api.Assertions.*;

class JwtTokenProviderTest {

    private JwtTokenProvider tokenProvider;

    private static final String SECRET = "test-secret-key-for-testing-only-must-be-long-enough-for-hmac-sha";
    private static final long EXPIRATION_MS = 86400000;

    @BeforeEach
    void setUp() {
        tokenProvider = new JwtTokenProvider(SECRET, EXPIRATION_MS);
    }

    @Test
    void generateTokenReturnsNonNullToken() {
        UserDetails userDetails = User.builder()
                .username("testuser")
                .password("pass")
                .authorities("ROLE_USER")
                .build();

        Authentication auth = new UsernamePasswordAuthenticationToken(userDetails, null, Collections.emptyList());

        String token = tokenProvider.generateToken(auth);

        assertNotNull(token);
        assertFalse(token.isEmpty());
    }

    @Test
    void getUsernameFromTokenReturnsCorrectUsername() {
        UserDetails userDetails = User.builder()
                .username("alice")
                .password("pass")
                .authorities("ROLE_USER")
                .build();

        Authentication auth = new UsernamePasswordAuthenticationToken(userDetails, null, Collections.emptyList());
        String token = tokenProvider.generateToken(auth);

        String username = tokenProvider.getUsernameFromToken(token);

        assertEquals("alice", username);
    }

    @Test
    void validateTokenReturnsTrueForValidToken() {
        UserDetails userDetails = User.builder()
                .username("validuser")
                .password("pass")
                .authorities("ROLE_USER")
                .build();

        Authentication auth = new UsernamePasswordAuthenticationToken(userDetails, null, Collections.emptyList());
        String token = tokenProvider.generateToken(auth);

        assertTrue(tokenProvider.validateToken(token));
    }

    @Test
    void validateTokenReturnsFalseForExpiredToken() {
        JwtTokenProvider shortLived = new JwtTokenProvider(SECRET, -1);

        UserDetails userDetails = User.builder()
                .username("expireduser")
                .password("pass")
                .authorities("ROLE_USER")
                .build();

        Authentication auth = new UsernamePasswordAuthenticationToken(userDetails, null, Collections.emptyList());
        String token = shortLived.generateToken(auth);

        assertFalse(shortLived.validateToken(token));
    }

    @Test
    void validateTokenReturnsFalseForTamperedToken() {
        UserDetails userDetails = User.builder()
                .username("tampered")
                .password("pass")
                .authorities("ROLE_USER")
                .build();

        Authentication auth = new UsernamePasswordAuthenticationToken(userDetails, null, Collections.emptyList());
        String token = tokenProvider.generateToken(auth);

        String tampered = token.substring(0, token.length() - 5) + "XXXXX";

        assertFalse(tokenProvider.validateToken(tampered));
    }

    @Test
    void getUsernameFromTokenThrowsForInvalidToken() {
        assertThrows(Exception.class, () ->
                tokenProvider.getUsernameFromToken("completely-invalid-token"));
    }
}
