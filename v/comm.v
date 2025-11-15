module spi_device(
    input wire clk,
    input wire reset,
    input wire ck,
    input wire cs,
    input wire si,
    output reg so,
    input wire send__ready,
    output reg send__ack,
    input wire [7:0] write,
    output reg recv__ready,
    input wire recv__ack,
    output reg [7:0] read);

    /*
     * Locally managed data:
     */
    reg active;
    reg [3:0] count;
    reg [7:0] incoming;
    reg [7:0] outgoing;
    reg was;

    /*
     * Logic for this module:
     */
    always @(posedge clk)
      begin
        if (! reset)
          begin
            recv__ready <= #1 1'b0;
            send__ack <= #1 1'b0;
            active <= #1 1'b0;
          end
        else
          begin
            was <= #1 ck;
            if (active & ! ck)
              begin
                so <= #1 outgoing[7];
              end
            if (! active)
              begin
                count <= #1 4'H0;
                active <= #1 ! cs;
              end
            else if (cs)
              begin
                active <= #1 1'b0;
              end
            else if (ck & ! was)
              begin
                count <= #1 (count + 4'H1);
                incoming <= #1 { incoming[6:0], si };
              end
            else if (count == 0)
              begin
                if (send__ready)
                  begin
                    send__ack <= #1 1'b1;
                    outgoing <= #1 write;
                  end
              end
            else if (count == 8)
              begin
                count <= #1 4'H0;
                read <= #1 incoming;
                recv__ready <= #1 1'b1;
              end
            else if (was & ! ck)
              begin
                outgoing[0] <= #1 1'b0;
                outgoing[7:1] <= #1 outgoing[6:0];
              end
            else
              begin
                recv__ready <= #1 (recv__ready & ! recv__ack);
                send__ack <= #1 (send__ack & send__ready);
              end
          end
      end

endmodule /* spi_device */

module mirror(
    input wire clk,
    input wire reset,
    input wire recv__ready,
    output reg recv__ack,
    input wire [7:0] cmd,
    output reg sent__ready,
    input wire sent__ack,
    output reg [7:0] reply,
    output reg [7:0] state);

    /*
     * Locally managed data:
     */

    /*
     * Logic for this module:
     */
    always @(posedge clk)
      begin
        if (! reset)
          begin
            state <= #1 8'H0;
            sent__ready <= #1 1'b0;
            recv__ack <= #1 1'b0;
          end
        else if (recv__ack & ! recv__ready)
          begin
            recv__ack <= #1 1'b0;
          end
        else if (sent__ready)
          begin
            sent__ready <= #1 ! sent__ack;
          end
        else if (recv__ready)
          begin
            recv__ack <= #1 1'b1;
            if (! recv__ack)
              begin
                sent__ready <= #1 1'b1;
                state <= #1 cmd;
                reply <= #1 state;
              end
          end
      end

endmodule /* mirror */

module top(input wire  reset,
	   input wire  clk,
	   output wire led_r,
	   output wire led_g,
	   output wire led_b,
	   output wire ice_so,
	   input wire  ice_sck,
	   input wire  ice_ssn,
	   input wire  ice_si);

   reg sck;
   reg ssn;
   reg si;

   wire	      reply_ready;
   wire	      reply_ack;
   wire [7:0] reply;
   wire	      cmd_ready;
   wire	      cmd_ack;
   wire [7:0] cmd;
   wire [7:0] state;

   spi_device dev(clk,
		  reset,
		  sck,
		  ssn,
		  si,
		  ice_so,
		  reply_ready,
		  reply_ack,
		  reply,
		  cmd_ready,
		  cmd_ack,
		  cmd);

   mirror store(clk,
		reset,
		cmd_ready,
		cmd_ack,
		cmd,
		reply_ready,
		reply_ack,
		reply,
		state);

   assign led_r = !state[2];
   assign led_g = !state[1];
   assign led_b = !state[0];

   /* sample inputs at clock edges */
   always @(posedge clk)
     begin
	ssn <= ice_ssn;
	sck <= ice_sck;
	si <= ice_si;
     end

endmodule /* top */
