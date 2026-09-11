package datalens.config;

import org.junit.jupiter.api.Test;
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.web.cors.CorsConfigurationSource;
import org.springframework.web.cors.UrlBasedCorsConfigurationSource;

import static org.junit.jupiter.api.Assertions.*;

class SecurityConfigTest {

    @Test
    void corsConfigurationSourceReturnsUrlBasedSource() {
        SecurityConfig config = new SecurityConfig(null);
        CorsConfigurationSource source = config.corsConfigurationSource();

        assertNotNull(source);
        assertInstanceOf(UrlBasedCorsConfigurationSource.class, source);
    }

    @Test
    void corsConfigurationSourceIsDifferentInstanceEachCall() {
        SecurityConfig config = new SecurityConfig(null);
        CorsConfigurationSource first = config.corsConfigurationSource();
        CorsConfigurationSource second = config.corsConfigurationSource();

        assertNotNull(first);
        assertNotNull(second);
    }

    @Test
    void passwordEncoderReturnsBCrypt() {
        SecurityConfig config = new SecurityConfig(null);
        PasswordEncoder encoder = config.passwordEncoder();

        assertNotNull(encoder);
        assertInstanceOf(BCryptPasswordEncoder.class, encoder);

        String raw = "testpassword";
        String encoded = encoder.encode(raw);
        assertNotEquals(raw, encoded);
        assertTrue(encoder.matches(raw, encoded));
    }
}
