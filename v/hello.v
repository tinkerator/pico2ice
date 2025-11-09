module top #(
	     parameter TOP_BIT = 23
	     ) (
		input wire reset,
		input wire  clk,
		output wire led_r,
		output wire led_g,
		output wire led_b
		);

   reg [TOP_BIT:0] delay;

   assign led_r = 1;  // turn off red
   assign led_g = 1;  // turn off green
   assign led_b = delay[TOP_BIT];  // blink blue

   always @(posedge clk) begin
      if (!reset)
	begin
	   delay <= 0;
	end
      else
	begin
	   delay <= delay+1;
	end
   end

endmodule // top
