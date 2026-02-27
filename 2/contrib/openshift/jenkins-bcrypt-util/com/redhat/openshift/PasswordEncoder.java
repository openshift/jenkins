package com.redhat.openshift;

import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;

/**
 * Encodes a password using Spring Security's BCryptPasswordEncoder.
 *
 * Previously used org.mindrot.jbcrypt.BCrypt, but Jenkins 2.509+ and LTS 2.516.1 removed
 * the jbcrypt library from the WAR and now uses Spring Security's BCryptPasswordEncoder
 * implementation internally.
 *
 * The output format ($2a$10$...) is identical and backward-compatible
 * with hashes produced by the old jbcrypt library.
 *
 * ref: https://www.jenkins.io/changelog/2.509/
 *      https://www.jenkins.io/changelog/2.516.1/
 *
 * @author kunalmemane
 */

public class PasswordEncoder {
	public static void main(String[] args) {
		String password = args[0];
		BCryptPasswordEncoder encoder = new BCryptPasswordEncoder();
		String encodedPassword = encoder.encode(password);
		System.out.println(encodedPassword);
	}
}
