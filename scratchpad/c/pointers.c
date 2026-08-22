#include <stdio.h>

void print_int_bytes(int value) {
	unsigned char *byte_ptr = (unsigned char *)&value;

	printf("Bytes of %d: ", value);

	for (size_t i = 0; i < sizeof(int); i++) {
		printf("%02X ", byte_ptr[i]);
	}

	printf("\n");
}

void print_int_bits(int value) {
	printf("Bits of %d: ", value);

	// loop 32 times, once for each bit (from left to right)
	for (int i = 31; i >= 0; i--) {
		// shift the bit to the right and check if it is 1 or 0
		int bit = (value >> i) & 1;
		printf("%d", bit);

		// add a space after every 8 bits (1 byte) to make it easy to
		// read
		if (i % 8 == 0) {
			printf(" ");
		}
	}
	printf("\n");
}

int main() {
	int number = 2022;
	print_int_bytes(number);
	print_int_bits(number);

	// %zu is the format specifier for type size_t
	printf("an int uses %zu bytes of memory\n", sizeof(int));
	// That prints "4" for me, but can vary by system.
}
