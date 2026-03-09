package com.redhat.openshift;

import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;

/**
 * Checks bcrypt password against a raw password 
 * using Spring Security's BCryptPasswordEncoder.
 *
 * @author kunalmemane
 */

public class PasswordChecker {
	public static void main(String[] args) {
		String rawPassword = args[0];
		String encodedPassword = args[1];

		BCryptPasswordEncoder encoder = new BCryptPasswordEncoder();
		if (!encoder.matches(rawPassword, encodedPassword)) {
			System.out.println("Detected password environment variable change, Jenkins configuration must be updated...");
			System.exit(1);
		}
		System.exit(0);
	}
}
