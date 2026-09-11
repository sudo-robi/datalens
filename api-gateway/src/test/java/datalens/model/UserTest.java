package datalens.model;

import org.junit.jupiter.api.Test;

import java.lang.reflect.Method;
import java.time.LocalDateTime;

import static org.junit.jupiter.api.Assertions.*;

class UserTest {

    @Test
    void builderCreatesValidUserWithAllFields() {
        LocalDateTime now = LocalDateTime.now();
        User user = User.builder()
                .id(1L)
                .username("alice")
                .email("alice@example.com")
                .passwordHash("hashedpw")
                .createdAt(now)
                .build();

        assertEquals(1L, user.getId());
        assertEquals("alice", user.getUsername());
        assertEquals("alice@example.com", user.getEmail());
        assertEquals("hashedpw", user.getPasswordHash());
        assertEquals(now, user.getCreatedAt());
    }

    @Test
    void gettersAndSettersWork() {
        User user = new User();
        user.setId(42L);
        user.setUsername("bob");
        user.setEmail("bob@example.com");
        user.setPasswordHash("hash");

        assertEquals(42L, user.getId());
        assertEquals("bob", user.getUsername());
        assertEquals("bob@example.com", user.getEmail());
        assertEquals("hash", user.getPasswordHash());
    }

    @Test
    void prePersistSetsCreatedAt() throws Exception {
        User user = new User();
        LocalDateTime before = LocalDateTime.now();

        Method onCreate = User.class.getDeclaredMethod("onCreate");
        onCreate.setAccessible(true);
        onCreate.invoke(user);

        LocalDateTime after = LocalDateTime.now();
        assertNotNull(user.getCreatedAt());
        assertFalse(user.getCreatedAt().isBefore(before));
        assertFalse(user.getCreatedAt().isAfter(after));
    }

    @Test
    void noArgsConstructorCreatesEmptyUser() {
        User user = new User();
        assertNull(user.getId());
        assertNull(user.getUsername());
    }
}
